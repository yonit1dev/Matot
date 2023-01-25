package peerconnection

import (
	"bytes"
	"fmt"
	"go-torrent-client/peer"
	"log"
	"net"
	"time"
)

type PeerClient struct {
	Conn     net.Conn
	Choked   bool
	Pieces   PeerPiece
	peer     peer.Peer
	infoHash [20]byte
	peerId   [20]byte
}

func NewPeerClient(peer peer.Peer, infoHash, peerId [20]byte) (*PeerClient, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 10*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	h, err := handshakePeer(conn, infoHash, peerId)
	if err != nil {
		conn.Close()
		log.Fatal(err)
	}

	fmt.Println(h.Protocol)

	pieces, err := ParseBitfieldMsg(conn)
	if err != nil {
		log.Fatalf("Piece recieveing error. %s", err)
	}

	return &PeerClient{
		Conn:     conn,
		Choked:   true,
		Pieces:   pieces,
		peer:     peer,
		infoHash: infoHash,
		peerId:   peerId,
	}, nil
}

func handshakePeer(conn net.Conn, infoHash [20]byte, peerId [20]byte) (*Handshake, error) {
	handshakeReq := NewHandshake(infoHash, peerId)

	_, err := conn.Write(handshakeReq.BufferHandshake())
	if err != nil {
		log.Fatalf("Handshaking error. %s", err)
	}

	handshakeRespone, err := ParseHandshake(conn)

	if err != nil {
		return nil, err
	}
	if !bytes.Equal(handshakeRespone.InfoHash[:], infoHash[:]) {
		log.Fatalf("Expected infohash %x but got %x", handshakeRespone.InfoHash, infoHash)
	}
	return handshakeRespone, nil
}