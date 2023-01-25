package main

import (
	"crypto/sha1"
	"fmt"
	"go-torrent-client/downloader"
	"go-torrent-client/settings"
	"go-torrent-client/tracker"
	"log"
	"math/rand"
	"os"
)

var trackerClient *tracker.TrackerClient
var torrentConfig *settings.Settings

func init() {
	log.SetOutput(os.Stdout)

	randomBytes := make([]byte, 20)
	rand.Read(randomBytes)

	torrentConfig = &settings.Settings{
		TransactionId: rand.Uint32(),
		PeerId:        sha1.Sum(randomBytes),
		Port:          uint16(6882),
	}
}

func main() {
	fmt.Println("Minimal Go Torrent Client!")

	fileReader, err := os.Open("./debian-mac-11.6.0-amd64-netinst.iso.torrent")
	if err != nil {
		log.Fatal("failed to open torrent file.")
	}
	defer fileReader.Close()

	t, err := tracker.OpenTorrent(fileReader)
	if err != nil {
		log.Fatal(err)
	}

	trackerClient = tracker.NewTrackerClient(&t)

	peers, err := trackerClient.GetPeersTCP(torrentConfig.PeerId, torrentConfig.Port)
	if err != nil {
		log.Fatal(err)
	}

	sentTorrent := downloader.Torrent{
		Peers:       peers,
		PeerID:      torrentConfig.PeerId,
		InfoHash:    t.InfoHash,
		PieceHashes: t.Pieces,
		PieceLength: int(t.PieceLength),
		Length:      int(t.Length),
		Name:        t.Name,
	}

	buffer, err := sentTorrent.Download()
	if err != nil {
		log.Fatal(err)
	}

	outFile, err := os.Create("./")
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()
	_, err = outFile.Write(buffer)
	if err != nil {
		log.Fatal(err)
	}

}
