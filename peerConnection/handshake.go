package peerconnection

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const PROTOCOL = "BitTorrent protocol"

type Handshake struct {
	Protocol string
	InfoHash [20]byte
	PeerId   [20]byte
}

func NewHandshake(infoHash, peerID [20]byte) *Handshake {
	return &Handshake{
		Protocol: PROTOCOL,
		InfoHash: infoHash,
		PeerId:   peerID,
	}
}

func (h *Handshake) BufferHandshake() []byte {
	buffer := bytes.NewBuffer(make([]byte, 0, 68))

	// protocol string length
	binary.Write(buffer, binary.BigEndian, uint8(19))
	// protocol string
	binary.Write(buffer, binary.BigEndian, []byte(h.Protocol))
	// reserved 8 bytes
	binary.Write(buffer, binary.BigEndian, uint32(0))
	binary.Write(buffer, binary.BigEndian, uint32(0))
	// infoHash
	binary.Write(buffer, binary.BigEndian, h.InfoHash)
	// peerId
	binary.Write(buffer, binary.BigEndian, h.PeerId)

	return buffer.Bytes()
}

func ParseHandshake(r io.Reader) (*Handshake, error) {
	handshakeBuffer := make([]byte, 1)
	_, err := io.ReadFull(r, handshakeBuffer)
	if err != nil {
		return nil, err
	}
	protocolLength := int(handshakeBuffer[0])

	if protocolLength == 0 {
		err := fmt.Errorf("protocol cannot be 0")
		return nil, err
	}

	parsedHandshake := make([]byte, 48+protocolLength)
	_, err = io.ReadFull(r, handshakeBuffer)
	if err != nil {
		return nil, err
	}

	var infoHash, peerID [20]byte

	copy(infoHash[:], parsedHandshake[protocolLength+8:protocolLength+8+20])
	copy(peerID[:], parsedHandshake[protocolLength+8+20:])

	h := Handshake{
		Protocol: string(parsedHandshake[0:protocolLength]),
		InfoHash: infoHash,
		PeerId:   peerID,
	}

	return &h, nil
}
