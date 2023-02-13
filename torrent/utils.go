package torrent

import (
	"log"
	"os"
)

func ToTrackerFile(m *Meta, src *os.File) (*TorrentFile, error) {
	length, err := m.Info.Size()
	if err != nil {
		log.Fatalf(err.Error())
	}

	hash := m.HashedInfo(src)

	pieceHashes, err := m.Info.splitPieces()
	if err != nil {
		log.Fatalf(err.Error())
	}

	torrent := &TorrentFile{
		Announce:    m.Announce,
		InfoHash:    hash,
		Name:        m.Info.Name,
		Length:      length,
		Pieces:      pieceHashes,
		PieceLength: m.Info.PieceLength,
	}

	return torrent, nil
}

func SaveTorrent(path string, buffer []byte) error {
	saved, err := os.Create(path)
	if err != nil {
		return err
	}
	defer saved.Close()

	_, err = saved.Write(buffer)
	if err != nil {
		return err
	}

	return nil
}
