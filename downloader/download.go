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

type Torrent struct {
	Peers       []peer.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

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

func attemptDownloadPiece(client *peerconnection.PeerClient, pw *pieceDwWork) ([]byte, error) {
	state := pieceDwProgress{
		index:      pw.index,
		peerClient: *client,
		buffer:     make([]byte, pw.length),
	}

	// Setting a deadline helps get unresponsive peers unstuck.
	// 30 seconds is more than enough time to download a 262 KB piece
	client.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer client.Conn.SetDeadline(time.Time{}) // Disable the deadline

	for state.downloaded < pw.length {
		// If unchoked, send requests until we have enough unfulfilled requests
		if !state.peerClient.Choked {
			for state.reqBacklog < PieceRequestBacklog && state.requested < pw.length {
				blockSize := ReqMsgSize
				// Last block might be shorter than the typical block
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

func (t *Torrent) startDownloadWorker(peer peer.Peer, workQueue chan *pieceDwWork, results chan *pieceDwResult) {
	c, err := peerconnection.NewPeerClient(peer, t.PeerID, t.InfoHash)
	if err != nil {
		log.Printf("Could not handshake with %s. Disconnecting\n", peer.IP)
		return
	}
	defer c.Conn.Close()
	log.Printf("Completed handshake with %s\n", peer.IP)

	c.SendUnchokeMsg()
	c.SendNotInteresetedMsg()

	for pw := range workQueue {
		if !c.Pieces.CheckPiece(pw.index) {
			workQueue <- pw // Put piece back on the queue
			continue
		}

		// Download the piece
		buf, err := attemptDownloadPiece(c, pw)
		if err != nil {
			log.Println("Exiting", err)
			workQueue <- pw // Put piece back on the queue
			return
		}

		err = CheckPiece(buf, pw.pieceHash)
		if err != nil {
			log.Printf("Piece #%d failed integrity check\n", pw.index)
			workQueue <- pw // Put piece back on the queue
			continue
		}

		c.SendHaveMsg(pw.index)
		results <- &pieceDwResult{pw.index, buf}
	}
}

func (t *Torrent) Download() ([]byte, error) {
	log.Println("Starting download for", t.Name)
	// Init queues for workers to retrieve work and send results
	workQueue := make(chan *pieceDwWork, len(t.PieceHashes))
	results := make(chan *pieceDwResult)
	for index, hash := range t.PieceHashes {
		length := CalcPieceSize(t.Length, t.PieceLength, index)
		workQueue <- &pieceDwWork{index, length, hash}
	}

	// Start workers
	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, workQueue, results)
	}

	// Collect results into a buffer until full
	buf := make([]byte, t.Length)
	donePieces := 0
	for donePieces < len(t.PieceHashes) {
		res := <-results
		begin, end := CalcPieceBounds(t.Length, t.PieceLength, res.index)
		copy(buf[begin:end], res.buffer)
		donePieces++

		percent := float64(donePieces) / float64(len(t.PieceHashes)) * 100
		numWorkers := runtime.NumGoroutine() - 1 // subtract 1 for main thread
		log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, res.index, numWorkers)
	}
	close(workQueue)

	return buf, nil
}
