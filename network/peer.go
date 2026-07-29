package network

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net"

	"github.com/alarakokbudak/dagitikSistemTransferPlatformu/crypto"
)

type Payload struct {
	Type string
	Data []byte
}

func SendPayload(conn net.Conn, p Payload, key []byte) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	encryptedData, err := crypto.Encrypt(data, key)
	if err != nil {
		return err
	}

	length := uint32(len(encryptedData))
	err = binary.Write(conn, binary.LittleEndian, length)
	if err != nil {
		return err
	}

	_, err = conn.Write(encryptedData)
	return err
}

func ReceivePayload(conn net.Conn, key []byte) (*Payload, error) {
	var length uint32
	err := binary.Read(conn, binary.LittleEndian, &length)
	if err != nil {
		return nil, err
	}

	encryptedData := make([]byte, length)
	_, err = io.ReadFull(conn, encryptedData)
	if err != nil {
		return nil, err
	}

	data, err := crypto.Decrypt(encryptedData, key)
	if err != nil {
		return nil, err
	}

	var p Payload
	err = json.Unmarshal(data, &p)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
