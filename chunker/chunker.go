package chunker

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

const (
	MinChunkSize    = 256 * 1024      // Minimum 256 KB
	TargetChunkSize = 1024 * 1024     // Hedef (Target) 1 MB
	MaxChunkSize    = 4 * 1024 * 1024 // Maksimum 4 MB
	WindowSize      = 64              // Kayan Özet Pencere Boyutu
	Mask            = TargetChunkSize - 1
)

type Chunk struct {
	Index  int
	Offset int64
	Size   int
	Hash   string
}

// ChunkFile splits the file into chunks using Content-Defined Chunking (CDC)
// with a Rolling Hash algorithm, solving the Byte Shifting problem.
func ChunkFile(filePath string) ([]Chunk, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var chunks []Chunk
	var index int
	var globalOffset int64

	// Sliding window buffer
	window := make([]byte, WindowSize)
	var windowPos int
	
	// Preallocate a buffer to hold the current chunk being built
	currentChunkBuf := make([]byte, 0, MaxChunkSize)

	var h uint32 // Rolling Hash state
	
	for {
		b, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		currentChunkBuf = append(currentChunkBuf, b)
		
		// Update the rolling hash state (Rabin-Karp / Adler style bitwise rotation)
		oldByte := window[windowPos]
		window[windowPos] = b
		windowPos = (windowPos + 1) % WindowSize
		
		// Very fast naive rolling hash calculation
		h = (h ^ uint32(oldByte))
		h = (h << 1) | (h >> 31) // Rotate left
		h = h ^ uint32(b)

		chunkSize := len(currentChunkBuf)
		
		// A chunk boundary is defined if the hash matches the mask (on average every TargetChunkSize bytes),
		// OR if the chunk reaches MaxChunkSize.
		isBoundary := (chunkSize >= MinChunkSize) && ((h & Mask) == 0)
		
		if isBoundary || chunkSize >= MaxChunkSize {
			// Chunk boundary found! Calculate true SHA-256 for the Merkle Tree
			hash := sha256.Sum256(currentChunkBuf)
			chunks = append(chunks, Chunk{
				Index:  index,
				Offset: globalOffset,
				Size:   chunkSize,
				Hash:   hex.EncodeToString(hash[:]),
			})
			globalOffset += int64(chunkSize)
			index++
			currentChunkBuf = currentChunkBuf[:0] // Reset buffer efficiently
			h = 0
		}
	}

	// Flush any remaining data as the last chunk
	if len(currentChunkBuf) > 0 {
		hash := sha256.Sum256(currentChunkBuf)
		chunks = append(chunks, Chunk{
			Index:  index,
			Offset: globalOffset,
			Size:   len(currentChunkBuf),
			Hash:   hex.EncodeToString(hash[:]),
		})
	}

	return chunks, nil
}
