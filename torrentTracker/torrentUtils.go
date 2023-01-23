package torrenttracker

import (
	"fmt"
	"os"

	"github.com/jackpal/bencode-go"
)

func OpenTorrentFile(filePath string) (TorrentFile, error) {
	torrentFile, err := os.Open(filePath)

	if err != nil {
		return TorrentFile{}, fmt.Errorf("couldn't open torrent file! Error: %s", err)
	}
	// close file no matter the output of the function
	defer torrentFile.Close()

	torrent := Torrent{}
	err = bencode.Unmarshal(torrentFile, &torrent)
	if err != nil {
		return TorrentFile{}, fmt.Errorf("couldn't decode torrent file! Error: %s", err)
	}

	return torrent.ConvTorrentFile()
}
