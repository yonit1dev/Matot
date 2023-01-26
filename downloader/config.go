package downloader

import (
	peerconnection "go-torrent-client/peerConnection"
	"log"
	"time"
)

// 16KB as specified in BEP
const ReqMsgSize = 16384

// recommended to be 5 in BitTorrent Paper
const PieceRequestBacklog = 5

// handling the downloading of pieces using go routines
// after a piece is downloaded we notify other go routines using channel (share memory by communicating)

// defining a piece download progress and a go routine work progress

// monitors the downloading of a piece by multiple go routines connecting to a peer
type pieceDwProgress struct {
	peerClient peerconnection.PeerClient
	buffer     []byte
	index      int
	recieved   int
	requested  int
	reqBacklog int
}

// holds the actions for downloading a pice
type pieceDwWork struct {
	index     int
	length    int
	pieceHash [20]byte
}

// holds the result of downloading a piece
type pieceDwResult struct {
	index  int
	buffer []byte
}

func (state *pieceDwProgress) readPeerConnectionMsg() error {
	msg, err := peerconnection.ParseMsg(state.peerClient.Conn)
	if err != nil {
		return err
	}

	// Keep Alive Msg returned
	if msg == nil {
		return nil
	}

	switch msg.ID {
	// recieve an unchoke message
	case peerconnection.UnChoke:
		state.peerClient.Choked = false
	// recieve a choke message
	case peerconnection.Choke:
		state.peerClient.Choked = true
	// recieve a have message
	case peerconnection.Have:
		// check for index of piece
		index, err := peerconnection.ParseHaveMsg(msg)
		if err != nil {
			return err
		}
		// recognize the bit for the piece
		state.peerClient.Pieces.SetBitPiece(index)
	// recieve a bitfield message
	case peerconnection.Bitfield:
		n, err := peerconnection.ParsePieceMsg(msg, state.index, state.buffer)
		if err != nil {
			return err
		}
		state.recieved += n
		state.reqBacklog--
	}
	return nil
}

func recievePiece(client *peerconnection.PeerClient, pw *pieceDwWork) ([]byte, error) {
	state := pieceDwProgress{
		index:      pw.index,
		peerClient: *client,
		buffer:     make([]byte, pw.length),
	}

	client.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer client.Conn.SetDeadline(time.Time{})

	// the amount we recieved is less than the length of the piece continue
	for state.recieved < pw.length {

		if !state.peerClient.Choked {
			for state.reqBacklog < PieceRequestBacklog && state.requested < pw.length {
				// pieces are split into blocks to request at once
				blockSize := ReqMsgSize

				if pw.length-state.requested < blockSize {
					blockSize = pw.length - state.requested
				}

				err := state.peerClient.SendRequestMsg(pw.index, state.requested, blockSize)
				if err != nil {
					log.Fatal(err)
				}
				state.reqBacklog++
				state.requested += blockSize
			}
		}

		err := state.readPeerConnectionMsg()
		if err != nil {
			return nil, err
		}
	}

	return state.buffer, nil
}
