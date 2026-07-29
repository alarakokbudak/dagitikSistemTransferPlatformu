package network

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/alarakokbudak/dagitikSistemTransferPlatformu/rsync"
)

type SyncManager struct {
	DirPath string
	Key     []byte
}

func NewSyncManager(dirPath string, key []byte) *SyncManager {
	return &SyncManager{
		DirPath: dirPath,
		Key:     key,
	}
}

func (s *SyncManager) StartServer(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()
	fmt.Println("Server listening on", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go s.handleConnection(conn)
	}
}

// Client logic (Sender with NEW data)
func (s *SyncManager) ConnectAndSync(address string, targetFilename string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	filePath := targetFilename
	fullPath := filepath.Join(s.DirPath, filePath)

	// Step 1: Ask server for its signatures of "rapor.txt"
	fmt.Printf("Requesting signature map for %s...\n", filePath)
	SendPayload(conn, Payload{Type: "SIG_REQ", Data: []byte(filePath)}, s.Key)

	// Step 2: Receive signature map
	sigPayload, _ := ReceivePayload(conn, s.Key)
	var signatures []rsync.BlockSignature
	json.Unmarshal(sigPayload.Data, &signatures)
	fmt.Printf("Received %d block signatures from Server.\n", len(signatures))

	// Step 3: Sliding window scan to calculate delta
	fmt.Println("Scanning file with sliding window to find patches...")
	operations, err := rsync.CalculateDelta(fullPath, signatures)
	if err != nil {
		return err
	}

	// Step 4: Send the patching instructions (Delta)
	opsData, _ := json.Marshal(operations)
	fmt.Printf("Sending %d patch instructions to Server...\n", len(operations))
	SendPayload(conn, Payload{Type: "DELTA", Data: opsData}, s.Key)

	return nil
}

// Server logic (Receiver with OLD data)
func (s *SyncManager) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Step 1: Wait for signature request
	reqPayload, _ := ReceivePayload(conn, s.Key)
	filePath := string(reqPayload.Data)
	fullPath := filepath.Join(s.DirPath, filePath)

	// Step 2: Generate Block Signatures
	fmt.Printf("Generating block map for %s...\n", filePath)
	signatures, _ := rsync.GenerateSignatures(fullPath)
	
	sigData, _ := json.Marshal(signatures)
	SendPayload(conn, Payload{Type: "SIG_RES", Data: sigData}, s.Key)

	// Step 3: Receive patching instructions (Delta)
	deltaPayload, _ := ReceivePayload(conn, s.Key)
	var operations []rsync.Operation
	json.Unmarshal(deltaPayload.Data, &operations)
	
	// Step 4: Apply Patch
	fmt.Printf("Received %d patch instructions. Patching file...\n", len(operations))
	tempPath := fullPath + ".tmp"
	err := rsync.ApplyPatch(fullPath, operations, tempPath)
	if err == nil {
		os.Rename(tempPath, fullPath)
		fmt.Println("File successfully patched using Rsync Delta-Sync!")
	} else {
		fmt.Println("Patch error:", err)
	}
}
