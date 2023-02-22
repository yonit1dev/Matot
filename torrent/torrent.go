package torrent

import (
	"crypto/sha1"
	"io"
	"log"
	"os"

	bencode1 "github.com/zeebo/bencode"
)

type InfoExtractor struct {
	RawInfo bencode1.RawMessage `bencode:"info"`
}

type File struct {
	Length uint64 `bencode:"length"`
	Path   string `bencode:"path"`
}

// meta file info key
type MetaInfo struct {
	Files       []File `bencode:"files"`
	Name        string `bencode:"name"`
	Length      uint64 `bencode:"length"`
	Pieces      string `bencode:"pieces"`
	PieceLength uint64 `bencode:"piece length"`
}

// meta file
type Meta struct {
	Announce string   `bencode:"announce"`
	Info     MetaInfo `bencode:"info"`
}

// torrent file to send tracker
type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	Name        string
	Length      uint64
	Pieces      [][20]byte
	PieceLength uint64
	ResumeFile  *os.File
}

func (metaInfo *MetaInfo) Size() (uint64, error) {
	var length uint64

	if len(metaInfo.Files) == 0 {
		length = metaInfo.Length
	} else {
		for _, item := range metaInfo.Files {
			length += item.Length
		}
	}

	return length, nil
}

func (meta *Meta) HashedInfo(src *os.File) (hashed [20]byte) {
	var rawInfo InfoExtractor
	// go back to the begining of the file
	src.Seek(0, 0)
	content, _ := io.ReadAll(src)

	err := bencode1.DecodeBytes(content, &rawInfo)
	if err != nil {
		log.Fatalf("Couldn't decode infohash bytes: %s", err)
	}

	hasher := sha1.New()
	hasher.Write(rawInfo.RawInfo)
	copy(hashed[:], hasher.Sum(nil))
	return
}

func (metaInfo *MetaInfo) splitPieces() ([][20]byte, error) {
	pieceLength := 20 // 20 bytes size of one piece hash
	pieceBuffer := []byte(metaInfo.Pieces)

	hashNum := len(pieceBuffer) / pieceLength
	pieceHashes := make([][20]byte, hashNum) // array containing the piece hashes of size 20 byte

	for i := 0; i < hashNum; i++ {
		copy(pieceHashes[i][:], pieceBuffer[i*pieceLength:(i+1)*pieceLength])
	}
	return pieceHashes, nil
}
