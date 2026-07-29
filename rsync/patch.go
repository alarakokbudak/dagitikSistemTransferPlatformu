package rsync

import (
	"fmt"
	"io"
	"os"
)

// ApplyPatch takes the old file, a list of patch operations (Delta),
// and writes the result to the destination file.
func ApplyPatch(oldFilePath string, operations []Operation, destFilePath string) error {
	oldFile, err := os.Open(oldFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// If old file doesn't exist, we can still patch if all ops are DATA
		oldFile = nil
	} else {
		defer oldFile.Close()
	}

	destFile, err := os.Create(destFilePath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	for _, op := range operations {
		if op.Type == OpMatch {
			if oldFile == nil {
				return fmt.Errorf("MATCH operation received but old file does not exist")
			}
			
			// A MATCH operation refers to a BlockIndex in the old file
			offset := int64(op.BlockIndex) * int64(BlockSize)
			
			// Seek to the start of the matching block
			_, err := oldFile.Seek(offset, io.SeekStart)
			if err != nil {
				return err
			}
			
			// Read the block
			blockData := make([]byte, BlockSize)
			bytesRead, err := io.ReadFull(oldFile, blockData)
			if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
				return err
			}
			
			// Write the block to the destination
			_, err = destFile.Write(blockData[:bytesRead])
			if err != nil {
				return err
			}
			
		} else if op.Type == OpData {
			// Literal DATA, just write it
			_, err = destFile.Write(op.Data)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unknown operation type: %s", op.Type)
		}
	}

	return nil
}
