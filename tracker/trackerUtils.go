package tracker

import (
	"go-torrent-client/torrent"
	"io"
	"log"

	"github.com/jackpal/bencode-go"
)

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	Name        string
	Length      uint64
	PieceLength uint64
	Pieces      [][20]byte
}

func OpenTorrent(r io.Reader) (TorrentFile, error) {
	t := torrent.Torrent{}
	err := bencode.Unmarshal(r, &t)
	if err != nil {
		log.Fatalf("failed to read torrent file. Error: %s", err)
	}

	return toTrackerTFile(&t)

}

func toTrackerTFile(t *torrent.Torrent) (TorrentFile, error) {
	infoHash, err := t.Info.CalcInfoHash()
	if err != nil {
		log.Fatalf(err.Error())
	}

	pieceHashes, err := t.Info.SplitPieces()
	if err != nil {
		log.Fatalf(err.Error())
	}

	tFile := TorrentFile{
		Announce:    t.Announce,
		InfoHash:    infoHash,
		Name:        t.Info.Name,
		Length:      t.Info.Length,
		PieceLength: t.Info.PieceLength,
		Pieces:      pieceHashes,
	}

	return tFile, nil

}
