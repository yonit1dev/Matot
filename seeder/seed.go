package seeder

import (
	"fmt"
	"log"
	"math"
	"matot/config"
	peerconnect "matot/peerConnect"
	"matot/torrent"
	"net"
	"os"
)

func SendTorrentPiece(conn net.Conn, t *torrent.TorrentFile, m *peerconnect.Message) error {
	req := RecieveReqMsg(m, t)
	log.Printf("Sending piece #%d", req.SpMsg.Index)

	tFile, err := os.Open(t.Name)
	if err != nil {
		log.Print("Couldn't open torrent")
		return err
	}
	defer tFile.Close()

	piece := make([]byte, req.BlockSize)
	_, err = tFile.ReadAt(piece, int64(req.SpMsg.Begin))

	if err != nil {
		log.Printf("Reading piece failed: %s", err)
		return err
	}

	log.Printf("Sending piece #%d", req.SpMsg.Index)
	msg := SendPiece(req, piece)

	conn.Write(msg.BufferMsg())

	return nil
}

func handleUploadConnection(t *torrent.TorrentFile, b peerconnect.BitFieldType, conn net.Conn) {
	defer conn.Close()

	handshake, err := peerconnect.CheckProtocol(&conn)
	if err != nil {
		log.Print("Failed handshake with peer")
		return
	}

	if handshake.InfoHash != t.InfoHash {
		log.Print("Don't have the specified infohash")
		return
	}

	// Sending handshake
	conn.Write(handshake.HandshakeMsg())

	// Sending bitfield
	pieces := b
	bitfieldMsg := peerconnect.Message{ID: peerconnect.BitField, Payload: pieces}
	conn.Write(bitfieldMsg.BufferMsg())

	_, err = peerconnect.ReadMsg(conn)
	if err != nil {
		return
	}

	msg := peerconnect.UnchokeMsg()

	_, err = conn.Write(msg)
	if err != nil {
		log.Print("Unchoke handler error")
		return
	}

	for {
		reqMsg, err := peerconnect.ReadMsg(conn)
		if err != nil {
			log.Print("Read message error", reqMsg, err)
			return
		}
		go SendTorrentPiece(conn, t, reqMsg)
	}
}

func UploadServer(t *torrent.TorrentFile, config *config.Config) {

	port := config.Port

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		log.Fatalf("Couldn't listen to connections! %s", err)
	}

	bSize := math.Ceil((float64(len(t.Pieces)) / float64(8)))
	b := make(peerconnect.BitFieldType, int(bSize))

	log.Printf("Server listening on 127.0.0.1:%d", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf("Couldn't accept connections! %s", err)
		}
		log.Printf("Connected with: %s", conn.RemoteAddr().String())
		go handleUploadConnection(t, b, conn)
	}
}
