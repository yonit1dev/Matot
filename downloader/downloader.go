package downloader

import (
	"fmt"
	peerconnect "goAssignment/peerConnect"
	"goAssignment/tracker.go"
	"log"
	"runtime"
	"time"
)

const ReqBacklog = 10
const BlockSize = 16384 // 16KB

type pieceDw struct {
	pieceHash [20]byte
	index     int
	length    int
}

type pieceDwResult struct {
	index  int
	result []byte
}

type workerProg struct {
	index      int
	peerClient *peerconnect.PeerConnection
	downloaded int
	requested  int
	buffer     []byte
	backlog    int
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
			return err
		}
		progress.peerClient.Bitfield.ChangeBit(index)
	case peerconnect.Piece:
		recieved, err := peerconnect.RecievePieceMsg(progress.index, progress.buffer, m)
		if err != nil {
			return err
		}
		progress.downloaded += recieved
		progress.backlog -= 1
	}

	return nil
}

func HandleConnection(peer tracker.Peer, infoHash, peerID [20]byte, dwQueue chan *pieceDw, dwResult chan *pieceDwResult) {
	c, err := peerconnect.NewPeerConnection(peer, infoHash, peerID)
	if err != nil {
		log.Print(err)
		return
	}

	defer c.Conn.Close()

	log.Printf("Completed Handshake with: %s", peer.String())

	c.SendInteresetedMsg()
	_, err = peerconnect.ReadMsg(c.Conn)
	if err != nil {
		log.Print(err)
		return
	}

	for pdw := range dwQueue {
		if !c.Bitfield.PieceExist(pdw.index) {
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

func downloadPiece(c *peerconnect.PeerConnection, pdw *pieceDw) ([]byte, error) {
	resultBuffer := make([]byte, pdw.length)

	progress := workerProg{
		index:      pdw.index,
		peerClient: c,
		buffer:     resultBuffer,
	}

	c.Conn.SetDeadline(time.Now().Add(60 * time.Second))
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

func DownloadT(pieceHashes [][20]byte, pieceLength int, length uint64, peerAdd []tracker.Peer, infoHash, peerID [20]byte) ([]byte, error) {
	fmt.Println("Starting torrent download")

	dwQueue := make(chan *pieceDw, len(pieceHashes))
	dwResults := make(chan *pieceDwResult)

	for index, hash := range pieceHashes {
		length := calcPieceSize(int(length), pieceLength, index)
		dwQueue <- &pieceDw{hash, index, length}
	}

	for _, peer := range peerAdd {
		go HandleConnection(peer, infoHash, peerID, dwQueue, dwResults)
	}

	resultBuffer := make([]byte, length)
	downloadedPieces := 0

	for downloadedPieces < len(pieceHashes) {
		result := <-dwResults
		begin, end := calcPieceBounds(int(length), pieceLength, result.index)
		copy(resultBuffer[begin:end], result.result)
		downloadedPieces += 1

		completed := float32(downloadedPieces) / float32(len(pieceHashes)) * 100

		numConnPeers := runtime.NumGoroutine() - 1

		fmt.Printf("(%0.2f%%) Finished piece %d from %d active peers\n", completed, result.index, numConnPeers)
	}
	close(dwQueue)

	return resultBuffer, nil
}
