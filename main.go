package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"math/rand"
	"matot/config"
	"matot/downloader"
	"matot/torrent"
	"matot/tracker.go"
	"os"

	"github.com/jackpal/bencode-go"
)

var torrentConfig *config.Config

func init() {
	log.SetOutput(os.Stdout)

	randomBytes := make([]byte, 20)
	rand.Read(randomBytes)

	torrentConfig = &config.Config{
		TransactionID: rand.Uint32(),
		PeerID:        sha1.Sum(randomBytes),
		Port:          uint16(6882),
	}
}

func main() {
	fmt.Println("GoTorrent Client")

	fileReader, err := os.Open("./samples/debian.iso.torrent")
	if err != nil {
		log.Fatalf("Couldn't open meta-file")
		return
	}
	defer fileReader.Close()

	meta := torrent.Meta{}
	err = bencode.Unmarshal(fileReader, &meta)

	if err != nil {
		log.Fatalf("Couldn't unmarshal meta-file")
		return
	}

	tf, err := torrent.ToTrackerFile(&meta, fileReader)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}

	fmt.Println(tf.InfoHash)

	//client
	client := tracker.NewTrackerClient(tf)

	interval, peerAdd := client.GetPeersTCP(torrentConfig)
	if err != nil {
		log.Print(err.Error())
		return
	}

	fmt.Println(interval)

	results, err := downloader.DownloadT(tf.Pieces, int(tf.PieceLength), tf.Length, peerAdd, tf.InfoHash, torrentConfig.PeerID)
	if err != nil {
		log.Print(err)
		return
	}

	torrent.SaveTorrent("./output", results)

}
