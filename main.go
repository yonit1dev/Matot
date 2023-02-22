package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"math/rand"
	"matot/config"
	"matot/downloader"
	"matot/seeder"
	"matot/torrent"
	"matot/tracker"
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
	fmt.Println("Matot - BitTorrent Client")

	fileReader, err := os.Open("./samples/kali.iso.torrent")
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

	// udp tracker
	// udpClient := tracker.CreateClient(tf)
	// defer udpClient.Close()

	// connect := udpClient.ConnectTracker(torrentConfig)
	// if err = connect.ValidateConnectResponse(torrentConfig); err != nil {
	// 	fmt.Println(err)
	// }

	// announceTracker := udpClient.AnnounceTracker(tf, torrentConfig)
	// if announceTracker.Action == tracker.ErrorID {
	// 	log.Fatal("Failed to get list of peer from the tracker")
	// }

	// tcp tracker
	client := tracker.NewTrackerClient(tf)

	interval, peerAdd := client.GetPeersTCP(torrentConfig)
	if err != nil {
		log.Print(err.Error())
		return
	}
	fmt.Println(interval)

	// saved file
	saved, err := os.Create(tf.Name)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer saved.Close()
	tf.ResumeFile = saved

	// download start
	results, err := downloader.DownloadT(tf.Pieces, int(tf.PieceLength), tf.Length, peerAdd, tf.InfoHash, torrentConfig.PeerID, tf.ResumeFile)
	if err != nil {
		log.Print(err)
		return
	}

	torrent.SaveTorrent(tf.Name, results)

	// seeder start
	seeder.UploadServer(tf, torrentConfig)

}
