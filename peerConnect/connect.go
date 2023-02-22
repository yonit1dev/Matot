package peerconnect

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"matot/tracker"
	"net"
	"time"
)

type PeerConnection struct {
	Conn     net.Conn
	Address  tracker.Peer
	infoHash [20]byte
	peerID   [20]byte
	Bitfield BitFieldType
	Choked   bool
}

func performHandshake(conn net.Conn, infohash, peerID [20]byte) error {
	err := conn.SetDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		fmt.Print("SetDeadline failed")
		return nil
	}
	defer conn.SetDeadline(time.Time{})

	hMsg := NewHandshake(infohash, peerID)

	_, err = conn.Write(hMsg.HandshakeMsg())
	if err != nil {
		return err
	}

	h, err := CheckProtocol(&conn)
	if err != nil {
		return err
	}

	if !bytes.Equal(h.InfoHash[:], infohash[:]) {
		return errors.New("wrong info hash from peer")
	}

	return nil
}
func NewPeerConnection(peer tracker.Peer, infoHash, peerID [20]byte) (*PeerConnection, error) {
	conn, err := net.Dial("tcp", peer.String())
	if err != nil {
		return nil, err
	}

	// defer conn.Close()

	err = performHandshake(conn, infoHash, peerID)
	if err != nil {
		log.Printf("handshaking failed with peer: %s", peer.String())
		conn.Close()
		return nil, err
	}

	b, err := RecieveBitfieldMsg(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &PeerConnection{
		Conn:     conn,
		Address:  peer,
		infoHash: infoHash,
		peerID:   peerID,
		Bitfield: b,
		Choked:   true,
	}, nil
}

func (pc *PeerConnection) SendRequestMsg(sp *SpecialMsg) error {
	msg := RequestMsgPayload(sp)

	_, err := pc.Conn.Write(msg.BufferMsg())
	return err
}

func (pc *PeerConnection) SendHaveMsg(index uint32) error {
	msg := HaveMsgPayload(index)

	_, err := pc.Conn.Write(msg.BufferMsg())
	return err
}

func (pc *PeerConnection) SendInteresetedMsg() error {
	msg := InterestedMsg()

	_, err := pc.Conn.Write(msg)
	return err
}

func (pc *PeerConnection) SendNotInterestedMsg() error {
	msg := NotInterestedMsg()

	_, err := pc.Conn.Write(msg)
	return err
}

func (pc *PeerConnection) SendChokeMsg() error {
	msg := ChokeMsg()

	_, err := pc.Conn.Write(msg)
	return err
}

func (pc *PeerConnection) SendUnchokeMsg() error {
	msg := UnchokeMsg()

	_, err := pc.Conn.Write(msg)
	return err
}

func (pc *PeerConnection) ReadConnBuffer() (*Message, error) {
	msg, err := ReadMsg(pc.Conn)
	if err != nil {
		return nil, err
	}

	return msg, nil
}
