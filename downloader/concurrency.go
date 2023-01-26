package downloader

import (
	"go-torrent-client/peer"
	peerconnection "go-torrent-client/peerConnection"
	"log"
	"runtime"
)

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
