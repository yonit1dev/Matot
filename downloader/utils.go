package downloader

import (
	"bytes"
	"crypto/sha1"
	"errors"
)

// checks integrity of piece
func checkPiece(buffer []byte, pieceHash [20]byte) error {
	hash := sha1.Sum(buffer)
	if !bytes.Equal(hash[:], pieceHash[:]) {
		return errors.New("downloaded piece is corrupt")
	}

	return nil
}

// calculats a piece beginning and end
func calcPieceBounds(tLength, pieceLength, index int) (begin, end int) {
	// pieceLength is the same for all pieces
	begin = index * pieceLength
	end = begin + pieceLength
	if end > tLength {
		end = tLength
	}
	return begin, end
}

// calculates the size of a pieces hash
func calcPieceSize(tLength, pieceLength, index int) int {
	begin, end := calcPieceBounds(tLength, pieceLength, index)
	return end - begin
}
