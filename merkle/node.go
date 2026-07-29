package merkle

import (
	"crypto/sha256"
	"encoding/hex"
)

type Node struct {
	Hash     string
	IsLeaf   bool
	FilePath string
	Left     *Node
	Right    *Node
}

func NewLeafNode(filePath string, fileHash string) *Node {
	return &Node{
		Hash:     fileHash,
		IsLeaf:   true,
		FilePath: filePath,
	}
}

func NewInternalNode(left, right *Node) *Node {
	hasher := sha256.New()
	hasher.Write([]byte(left.Hash + right.Hash))
	hashString := hex.EncodeToString(hasher.Sum(nil))

	return &Node{
		Hash:   hashString,
		IsLeaf: false,
		Left:   left,
		Right:  right,
	}
}
