package downloader

import peerconnection "go-torrent-client/peerConnection"

// 16KB as specified in BEP
const ReqMsgSize = 16384

// recommended to be 5 in BitTorrent Paper
const PieceRequestBacklog = 5

// handling the downloading of pieces using go routines
// after a piece is downloaded we notify other go routines using channel (share memory by communicating)

// defining a piece download progress and a go routine work progress

type pieceDwProgress struct {
	peerClient peerconnection.PeerClient
	buffer     []byte
	index      int
	downloaded int
	requested  int
	reqBacklog int
}

type pieceDwWork struct {
	index     int
	length    int
	pieceHash [20]byte
}

type pieceDwResult struct {
	index  int
	buffer []byte
}

func (state *pieceDwProgress) readPeerConnectionMsg() error {
	msg, err := peerconnection.ParseMsg(state.peerClient.Conn) // this call blocks
	if err != nil {
		return err
	}

	// Keep Alive Msg returned
	if msg == nil {
		return nil
	}

	switch msg.ID {
	case peerconnection.UnChoke:
		state.peerClient.Choked = false
	case peerconnection.Choke:
		state.peerClient.Choked = true
	case peerconnection.Have:
		index, err := peerconnection.ParseHaveMsg(msg)
		if err != nil {
			return err
		}
		state.peerClient.Pieces.SetBitPiece(index)
	case peerconnection.Bitfield:
		n, err := peerconnection.ParsePieceMsg(msg, state.index, state.buffer)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.reqBacklog--
	}
	return nil
}
