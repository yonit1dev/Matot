package torrenttracker

import (
	"bytes"
	"crypto/sha1"
	"fmt"

	"github.com/jackpal/bencode-go"
)

type TorrentInfo struct {
	Files       []TorrentFile `bencode:"files"`
	Name        string        `bencode:"name"`
	Length      uint64        `bencode:"length"`
	Pieces      string        `bencode:"pieces"`
	PieceLength int           `bencode:"piece length"`
}

type Torrent struct {
	Announce string      `bencode:"announce"`
	Comment  string      `bencode:"comment"`
	Info     TorrentInfo `bencode:"info"`
}

type TorrentFile struct {
	Announce    string
	Name        string
	Length      uint64
	PieceLength int
	PieceHashes [][20]byte
	InfoHash    [20]byte
}

// calculate the infohash for pieces of the torrent
func (tInfo *TorrentInfo) calcInfoHash() ([20]byte, error) {
	var buffer bytes.Buffer

	// TODO: error handling
	err := bencode.Marshal(&buffer, *tInfo)
	if err != nil {
		return [20]byte{}, fmt.Errorf("InfoHash calculation error. Error: %s", err)
	}

	hash := sha1.Sum(buffer.Bytes())
	return hash, nil
}

// splits the piece hashes for convenient formatting
func (tInfo *TorrentInfo) calcPieceHashes() ([][20]byte, error) {
	hashLength := 20
	buffer := []byte(tInfo.Pieces)

	numHashes := len(buffer) / hashLength
	hash := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hash[i][:], buffer[i*hashLength:(i+1)*hashLength])
	}
	return hash, nil
}

func (t *Torrent) ConvTorrentFile() (TorrentFile, error) {
	infoHash, err := t.Info.calcInfoHash()

	if err != nil {
		return TorrentFile{}, fmt.Errorf("couldn't compute info hash. Error: %s", err)
	}

	pieceHashes, err := t.Info.calcPieceHashes()
	if err != nil {
		return TorrentFile{}, fmt.Errorf("couldn't compute hash for each piece. Error: %s", err)
	}

	tor := TorrentFile{
		Announce:    t.Announce,
		Name:        t.Info.Name,
		Length:      t.Info.Length,
		PieceLength: t.Info.PieceLength,
		PieceHashes: pieceHashes,
		InfoHash:    infoHash,
	}
	return tor, nil

}
