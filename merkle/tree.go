package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/alarakokbudak/dagitikSistemTransferPlatformu/chunker"
)

func ScanDirectory(dirPath string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			// Remove the dirPath prefix to make paths relative and comparable between peers
			relPath, err := filepath.Rel(dirPath, path)
			if err != nil {
				return err
			}
			// Use forward slashes for consistency across OS
			relPath = filepath.ToSlash(relPath)
			files = append(files, relPath)
		}
		return nil
	})
	return files, err
}

func BuildTreeForDirectory(dirPath string) (*Node, map[string][]chunker.Chunk, error) {
	files, err := ScanDirectory(dirPath)
	if err != nil {
		return nil, nil, err
	}

	if len(files) == 0 {
		return nil, nil, nil
	}

	sort.Strings(files)

	var leaves []*Node
	fileChunksMap := make(map[string][]chunker.Chunk)

	for _, file := range files {
		fullPath := filepath.Join(dirPath, filepath.FromSlash(file))
		chunks, err := chunker.ChunkFile(fullPath)
		if err != nil {
			return nil, nil, err
		}
		fileChunksMap[file] = chunks

		hasher := sha256.New()
		for _, c := range chunks {
			hasher.Write([]byte(c.Hash))
		}
		fileHash := hex.EncodeToString(hasher.Sum(nil))

		leaves = append(leaves, NewLeafNode(file, fileHash))
	}

	root := buildTree(leaves)
	return root, fileChunksMap, nil
}

func buildTree(nodes []*Node) *Node {
	if len(nodes) == 1 {
		return nodes[0]
	}

	var level []*Node
	for i := 0; i < len(nodes); i += 2 {
		if i+1 < len(nodes) {
			level = append(level, NewInternalNode(nodes[i], nodes[i+1]))
		} else {
			level = append(level, NewInternalNode(nodes[i], nodes[i]))
		}
	}

	return buildTree(level)
}

// CompareTrees returns a list of file paths that differ between two trees.
// In a real implementation, you'd traverse the tree to find differences efficiently.
// For simplicity in this demo, we'll return paths of differing leaves.
func CompareTrees(local, remote *Node) []string {
	var diff []string
	
	if local == nil && remote == nil {
		return diff
	}
	if local == nil || remote == nil || local.Hash != remote.Hash {
		// Traverse to find differences
		diff = append(diff, findDifferingLeaves(local, remote)...)
	}
	return diff
}

func findDifferingLeaves(local, remote *Node) []string {
	// A naive approach for the demo: if hashes differ, we eventually hit leaves.
	// We should just return the file paths of leaves that differ.
	// For simplicity, let's just collect all local leaves and remote leaves and compare hashes.
	// A robust implementation would do proper tree traversal.
	
	localLeaves := getLeaves(local)
	remoteLeaves := getLeaves(remote)
	
	var diff []string
	remoteMap := make(map[string]string)
	for _, l := range remoteLeaves {
		remoteMap[l.FilePath] = l.Hash
	}
	
	for _, l := range localLeaves {
		if rHash, ok := remoteMap[l.FilePath]; !ok || rHash != l.Hash {
			diff = append(diff, l.FilePath)
		}
	}
	
	for _, l := range remoteLeaves {
		if _, ok := remoteMap[l.FilePath]; !ok {
			// Remote has a file local doesn't have, or it differs.
			// To keep it simple, we just flag it.
			diff = append(diff, l.FilePath) // We might add duplicates, but we can deduplicate later.
		}
	}
	
	// Deduplicate
	dedup := make(map[string]bool)
	var finalDiff []string
	for _, f := range diff {
		if !dedup[f] {
			dedup[f] = true
			finalDiff = append(finalDiff, f)
		}
	}
	
	return finalDiff
}

func getLeaves(n *Node) []*Node {
	if n == nil {
		return nil
	}
	if n.IsLeaf {
		return []*Node{n}
	}
	var leaves []*Node
	leaves = append(leaves, getLeaves(n.Left)...)
	leaves = append(leaves, getLeaves(n.Right)...)
	return leaves
}
