package torrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"

	"github.com/jackpal/bencode-go"
)

type TorrentInfo struct {
	Name        string `bencode:"name"`
	Length      uint64 `bencode:"length"`
	Pieces      string `bencode:"pieces"`
	PieceLength uint64 `bencode:"piece length"`
}
type Torrent struct {
	Announce string      `bencode:"announce"`
	Info     TorrentInfo `bencode:"info"`
}

func (tInfo *TorrentInfo) CalcInfoHash() ([20]byte, error) {
	var infoHashBuf bytes.Buffer

	err := bencode.Marshal(&infoHashBuf, *tInfo)
	if err != nil {
		log.Fatalf("unable to calculate infohash. Error: %s", err)
	}

	infoHash := sha1.Sum(infoHashBuf.Bytes())

	return infoHash, nil

}

func (tInfo *TorrentInfo) SplitPieces() ([][20]byte, error) {
	pieceHashLength := 20 // Length of SHA-1 hash
	pieceBuffer := []byte(tInfo.Pieces)
	if len(pieceBuffer)%pieceHashLength != 0 {
		err := fmt.Errorf("received malformed pieces of length %d", len(pieceBuffer))
		return nil, err
	}
	numHashes := len(pieceBuffer) / pieceHashLength
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], pieceBuffer[i*pieceHashLength:(i+1)*pieceHashLength])
	}
	return hashes, nil
}
