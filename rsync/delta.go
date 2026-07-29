package rsync

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

const (
	OpMatch = "MATCH"
	OpData  = "DATA"
)

// Operation represents a delta instruction (Patch)
type Operation struct {
	Type       string // "MATCH" or "DATA"
	BlockIndex int    // Used if Type == MATCH
	Data       []byte // Used if Type == DATA
}

// CalculateDelta reads the new file, slides a window byte-by-byte,
// and generates a list of operations based on the old file's signatures.
func CalculateDelta(filePath string, signatures []BlockSignature) ([]Operation, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	weakMap := make(map[uint32][]BlockSignature)
	for _, sig := range signatures {
		weakMap[sig.WeakHash] = append(weakMap[sig.WeakHash], sig)
	}

	var ops []Operation
	var literalBuf []byte

	// If receiver has no file (empty signatures), just send the whole file as DATA
	if len(signatures) == 0 {
		data, err := io.ReadAll(reader)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if len(data) > 0 {
			ops = append(ops, Operation{Type: OpData, Data: data})
		}
		return ops, nil
	}

	window := make([]byte, BlockSize)
	bytesRead, err := io.ReadFull(reader, window)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, err
	}

	if bytesRead == 0 {
		return ops, nil
	}

	window = window[:bytesRead]
	weak := computeWeakHash(window)
	
	head := 0 // points to the oldest byte in the circular window

	for {
		matchIndex := -1
		
		if sigs, ok := weakMap[weak]; ok {
			// Construct flat window for SHA256 from circular buffer
			flatWindow := make([]byte, len(window))
			copy(flatWindow, window[head:])
			copy(flatWindow[len(window)-head:], window[:head])

			strongHashBytes := sha256.Sum256(flatWindow)
			strong := hex.EncodeToString(strongHashBytes[:])
			
			for _, sig := range sigs {
				if sig.StrongHash == strong {
					matchIndex = sig.Index
					break
				}
			}
		}

		if matchIndex != -1 {
			if len(literalBuf) > 0 {
				dataCopy := make([]byte, len(literalBuf))
				copy(dataCopy, literalBuf)
				ops = append(ops, Operation{Type: OpData, Data: dataCopy})
				literalBuf = literalBuf[:0]
			}

			ops = append(ops, Operation{Type: OpMatch, BlockIndex: matchIndex})
			
			bytesRead, err = io.ReadFull(reader, window[:cap(window)])
			if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
				return nil, err
			}
			
			if bytesRead == 0 {
				break
			}
			
			window = window[:bytesRead]
			weak = computeWeakHash(window)
			head = 0
		} else {
			oldByte := window[head]
			literalBuf = append(literalBuf, oldByte)
			
			newByte, err := reader.ReadByte()
			if err != nil {
				if err == io.EOF {
					// Flush the rest of the window
					for i := 1; i < len(window); i++ {
						idx := (head + i) % len(window)
						literalBuf = append(literalBuf, window[idx])
					}
					break
				}
				return nil, err
			}
			
			window[head] = newByte
			weak = RollWeakHash(weak, oldByte, newByte, len(window))
			
			head = (head + 1) % len(window)
		}
	}

	if len(literalBuf) > 0 {
		ops = append(ops, Operation{Type: OpData, Data: literalBuf})
	}

	return ops, nil
}
