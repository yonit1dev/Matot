package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"math/rand"
	"os"
	"torrentClient/settings"
	torrenttracker "torrentClient/torrentTracker"
)

var torrentConfig *settings.Settings
var torrentClient *torrenttracker.Client

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
	fmt.Println("Minimalist Go Torrent Client")

	torrent, err := torrenttracker.OpenTorrentFile("./sample.torrent")
	if err != nil {
		log.Fatal(err)
	}

	torrentClient = torrenttracker.CreateClient(&torrent)
	defer torrentClient.Close()

	connectTracker := torrentClient.ConnectTracker(&torrent, torrentConfig)

	if err = connectTracker.ValidateConnectResponse(torrentConfig); err != nil {
		fmt.Println(err)
	}

	// announceTracker := torrentClient.AnnounceTracker(&torrent, torrentConfig)

	// if announceTracker.Action == torrenttracker.ErrorID {
	// 	log.Fatal("Failed to get list of peer from the tracker")
	// }

	// fmt.Printf("Peers: %s", announceTracker.Addresses)

}
