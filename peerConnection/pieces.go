package peerconnection

import (
	"fmt"
	"net"
	"time"
)

// the piece that a peer has represented as a byte(bitfields)
type PeerPiece []byte

// if a bit is set to 0 in the peer piece byte representation then the peer doesn't have the piece
// if it's one, the peer has the piece
func (piece PeerPiece) CheckPiece(index int) bool {
	bytePos := index / 8
	slice := index % 8

	return piece[bytePos]>>(7-slice) != 0
}

func (piece PeerPiece) SetBitPiece(index int) {
	bytePos := index / 8
	slice := index % 8

	piece[bytePos] |= 1 << (7 - slice)
}

func ParseBitfieldMsg(conn net.Conn) (PeerPiece, error) {
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	msg, err := ParseMsg(conn)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		err := fmt.Errorf("expected pieces but got nil")
		return nil, err
	}
	if msg.ID != Bitfield {
		err := fmt.Errorf("expected pieces but got msgID %d", msg.ID)
		return nil, err
	}

	return msg.Payload, nil
}
