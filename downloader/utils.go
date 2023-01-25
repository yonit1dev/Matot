package downloader

import (
	"bytes"
	"crypto/sha1"
	"log"
)

func CheckPiece(buffer []byte, pieceHash [20]byte) error {
	hash := sha1.Sum(buffer)
	if !bytes.Equal(hash[:], pieceHash[:]) {
		log.Fatal("Downloaded piece is corrupt")
	}

	return nil
}

func CalcPieceBounds(tLength, pieceLength, index int) (begin, end int) {
	begin = index * pieceLength
	end = begin + pieceLength
	if end > tLength {
		end = tLength
	}
	return begin, end
}

func CalcPieceSize(tLength, pieceLength, index int) int {
	begin, end := CalcPieceBounds(tLength, pieceLength, index)
	return end - begin
}
