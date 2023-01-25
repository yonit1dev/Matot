package main

import (
	"crypto/sha1"
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
	// fmt.Println("Minimal Go Torrent Client!")

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

	trackerClient.GetPeersTCP(torrentConfig.PeerId, torrentConfig.Port)

	// client := tracker.CreateClient(&t)

	// connectTracker := client.ConnectTracker(torrentConfig)

	// if err = connectTracker.ValidateConnectResponse(torrentConfig); err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Println(connectTracker.Action)

	// announceTracker := client.AnnounceTracker(&t, torrentConfig)

	// fmt.Println(announceTracker.Interval)
	// fmt.Println(announceTracker.Addresses)

}
