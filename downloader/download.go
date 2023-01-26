package downloader

import (
	"go-torrent-client/peer"
	peerconnection "go-torrent-client/peerConnection"
	"log"
	"runtime"
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

func initDwThread(peer peer.Peer, peerId [20]byte, infoHash [20]byte, workQueue chan *pieceDwWork, results chan *pieceDwResult) {
	c, err := peerconnection.NewPeerClient(peer, peerId, infoHash)
	if err != nil {
		log.Fatalf("Handshake failed with peer: %s", peer.IP)
	}

	defer c.Conn.Close()
	log.Printf("Completed handshake with %s\n", peer.IP)

	c.SendUnchokeMsg()
	c.SendInteresetedMsg()

	for pw := range workQueue {
		if !c.Pieces.CheckPiece(pw.index) {
			workQueue <- pw
			continue
		}

		buf, err := recievePiece(c, pw)
		if err != nil {
			log.Println("Couldn't recieve piece. Done!", err)
			workQueue <- pw
			return
		}

		err = CheckPiece(buf, pw.pieceHash)
		if err != nil {
			log.Printf("Malformed piece: %d", pw.index)
			workQueue <- pw
			continue
		}

		c.SendHaveMsg(pw.index)
		results <- &pieceDwResult{pw.index, buf}
	}
}

func Download(peers []peer.Peer, peerId [20]byte, infoHash [20]byte, tLength int, pieceLength int, pieceHashes [][20]byte) ([]byte, error) {
	dwQueue := make(chan *pieceDwWork, len(pieceHashes))
	dwResults := make(chan *pieceDwResult)
	for index, hash := range pieceHashes {
		length := CalcPieceSize(tLength, pieceLength, index)
		dwQueue <- &pieceDwWork{index, length, hash}
	}

	for _, peer := range peers[0:5] {
		go initDwThread(peer, peerId, infoHash, dwQueue, dwResults)
	}

	buf := make([]byte, tLength)
	donePieces := 0
	for donePieces < len(pieceHashes) {
		res := <-dwResults
		begin, end := CalcPieceBounds(tLength, pieceLength, res.index)
		copy(buf[begin:end], res.buffer)
		donePieces += 1

		percent := float64(donePieces) / float64(len(pieceHashes)) * 100
		numWorkers := runtime.NumGoroutine() - 1
		log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, res.index, numWorkers)
	}
	close(dwQueue)

	return buf, nil
}
