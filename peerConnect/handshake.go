package peerconnect

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

type Handshake struct {
	Protocol string
	InfoHash [20]byte
	PeerID   [20]byte
}

func NewHandshake(infoHash, peerID [20]byte) *Handshake {
	return &Handshake{
		Protocol: "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

func (handshake *Handshake) HandshakeMsg() []byte {
	handshakeBufferSize := 68 // length of protocol string + infohash + peerid + reserved

	buffer := bytes.NewBuffer(make([]byte, 0, handshakeBufferSize))

	// protocol string length
	binary.Write(buffer, binary.BigEndian, uint8(19))
	// protocol string
	binary.Write(buffer, binary.BigEndian, []byte(handshake.Protocol))
	// reserved
	binary.Write(buffer, binary.BigEndian, uint32(0))
	binary.Write(buffer, binary.BigEndian, uint32(0))
	// infoHash
	binary.Write(buffer, binary.BigEndian, handshake.InfoHash)
	// peerId
	binary.Write(buffer, binary.BigEndian, handshake.PeerID)

	return buffer.Bytes()
}

func CheckProtocol(conn *net.Conn) (*Handshake, error) {
	// first sends a length prefix
	length := make([]byte, 1)
	message := make([]byte, 19)

	err := binary.Read(*conn, binary.BigEndian, length)
	if err != nil {
		return nil, err
	}
	err = binary.Read(*conn, binary.BigEndian, message)
	if err != nil {
		return nil, err
	}

	if string(message) != "BitTorrent protocol" {
		fmt.Println(string(message) != "BitTorrent Protocol")
		return nil, errors.New("bitTorrent HandShake Failed")
	}

	// return handshake msg
	hBufferLen := 48

	handshakeBuffer := make([]byte, hBufferLen)
	binary.Read(*conn, binary.BigEndian, handshakeBuffer[0:8])
	binary.Read(*conn, binary.BigEndian, handshakeBuffer[8:28])
	binary.Read(*conn, binary.BigEndian, handshakeBuffer[28:48])

	var infoHash, peerID [20]byte
	// reserved := uint(8)

	copy(infoHash[:], handshakeBuffer[8:28])
	copy(peerID[:], handshakeBuffer[28:48])

	h := Handshake{
		Protocol: string(message),
		InfoHash: infoHash,
		PeerID:   peerID,
	}

	return &h, nil

}
