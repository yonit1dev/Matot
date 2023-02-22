package downloader

import (
	"fmt"
	"log"
	peerconnect "matot/peerConnect"
	"matot/tracker"
	"os"
	"runtime"
	"time"
)

const ReqBacklog = 5    // 5 requests pending
const BlockSize = 16384 // 16KB

type pieceDw struct {
	pieceHash [20]byte // piecehash of each piece
	index     int      // position of each piece in the pieces array
	length    int      // length of the piece
}

type pieceDwResult struct {
	index  int    // downloaded piece index
	result []byte // downloaded bytes
}

// keeps track of a go routine download progress
type workerProg struct {
	index      int                         // index of piece
	peerClient *peerconnect.PeerConnection // connection with peer
	downloaded int                         // downloaded bytes length
	requested  int                         // requested bytes length
	buffer     []byte                      // result buffer to keep track of downloaded bytes
	backlog    int                         // requests pending
}

func (progress *workerProg) readMsg() error {
	m, err := progress.peerClient.ReadConnBuffer()
	if err != nil {
		fmt.Println(err)
		return err
	}

	if m == nil {
		return nil
	}

	switch m.ID {
	case peerconnect.Choke:
		progress.peerClient.Choked = true
	case peerconnect.Unchoke:
		progress.peerClient.Choked = false
	case peerconnect.Have:
		index, err := peerconnect.RecieveHaveMsg(m)
		if err != nil {
			log.Print(err)
			return err
		}
		progress.peerClient.Bitfield.ChangeBit(index)
	case peerconnect.Piece:
		recieved, err := peerconnect.RecievePieceMsg(progress.index, progress.buffer, m)
		if err != nil {
			log.Print(err)
			return err
		}
		progress.downloaded += recieved
		progress.backlog -= 1
	}

	return nil
}

func downloadPiece(c *peerconnect.PeerConnection, pdw *pieceDw) ([]byte, error) {
	resultBuffer := make([]byte, pdw.length)

	progress := workerProg{
		index:      pdw.index,
		peerClient: c,
		buffer:     resultBuffer,
	}

	c.Conn.SetDeadline(time.Now().Add(25 * time.Second))
	defer c.Conn.SetDeadline(time.Time{})

	for progress.downloaded < pdw.length {
		if !progress.peerClient.Choked {
			for progress.backlog < ReqBacklog && progress.requested < pdw.length {
				block := BlockSize

				if pdw.length-progress.requested < block {
					block = pdw.length - progress.requested
				}

				sp := peerconnect.SpecialMsg{
					Index:  uint32(pdw.index),
					Begin:  uint32(progress.requested),
					Length: uint32(block),
				}

				err := c.SendRequestMsg(&sp)
				if err != nil {
					return nil, err
				}

				progress.backlog += 1
				progress.requested += block
			}
		}

		err := progress.readMsg()
		if err != nil {
			return nil, err
		}
	}

	return progress.buffer, nil
}

func handleConnection(peer tracker.Peer, infoHash, peerID [20]byte, dwQueue chan *pieceDw, dwResult chan *pieceDwResult, b peerconnect.BitFieldType) {
	c, err := peerconnect.NewPeerConnection(peer, infoHash, peerID)
	if err != nil {
		log.Print(err)
		return
	}

	defer c.Conn.Close()

	// sending unchoke and intereseted message after verifying handshake
	err = c.SendUnchokeMsg()
	if err != nil {
		log.Print(err)
		return
	}
	err = c.SendInteresetedMsg()
	if err != nil {
		log.Print(err)
		return
	}

	for pdw := range dwQueue {
		if !c.Bitfield.PieceExist(pdw.index) && !b.PieceExist(pdw.index) {
			dwQueue <- pdw
			continue
		}

		resultBuffer, err := downloadPiece(c, pdw)
		if err != nil {
			log.Println("Couldn't download piece. Done!", err)
			dwQueue <- pdw
			return
		}

		err = checkPiece(resultBuffer, pdw.pieceHash)
		if err != nil {
			log.Printf("Malformed piece: %d", pdw.index)
			dwQueue <- pdw
			continue
		}

		c.SendHaveMsg(uint32(pdw.index))
		dwResult <- &pieceDwResult{pdw.index, resultBuffer}
	}
}

func DownloadT(pieceHashes [][20]byte, pieceLength int, length uint64, peerAdd []tracker.Peer, infoHash, peerID [20]byte, f *os.File) ([]byte, error) {
	fmt.Println("Starting torrent download")

	// channel that keeps track of pieces to download
	dwQueue := make(chan *pieceDw, len(pieceHashes))

	// channel that keeps track of downloaded pieces and their result
	dwResults := make(chan *pieceDwResult)

	// local pieces
	var b peerconnect.BitFieldType
	// downloaded pieces
	downloadedPieces := 0

	// index is the position of the hash in the pieceHash array
	for index, hash := range pieceHashes {
		length := calcPieceSize(int(length), pieceLength, index)

		begin, _ := calcPieceBounds(int(length), pieceLength, index)

		piece := make([]byte, length)
		_, err := f.ReadAt(piece, int64(begin))
		if err != nil {
			log.Printf("No local file: %s", err)
		}

		check := checkPiece(piece, hash)
		if check == nil {
			downloadedPieces += 1
			b.ChangeBit(index)
		} else {
			dwQueue <- &pieceDw{hash, index, length}
		}
	}

	// downloading pieces
	for _, peer := range peerAdd {
		go handleConnection(peer, infoHash, peerID, dwQueue, dwResults, b)
	}

	resultBuffer := make([]byte, length)

	for downloadedPieces < len(pieceHashes) {
		result := <-dwResults
		begin, end := calcPieceBounds(int(length), pieceLength, result.index)

		_, err := f.WriteAt(result.result, int64(begin))
		if err != nil {
			log.Print("Coudln't save piece locally")
		}

		copy(resultBuffer[begin:end], result.result)
		downloadedPieces += 1

		downloaded := float32(downloadedPieces) / float32(len(pieceHashes)) * 100

		numConnPeers := runtime.NumGoroutine() - 1

		fmt.Printf("Progress: (%0.2f%%). Downloading from %d active peers\n", downloaded, numConnPeers)
	}
	log.Print("Download done!")
	close(dwQueue)

	return resultBuffer, nil
}
