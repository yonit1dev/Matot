package torrenttracker

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	peerconnection "torrentClient/peerConnection"
	"torrentClient/settings"
)

const UDProtocol = uint64(0x41727101980)

const (
	ConnectID  = uint32(0)
	AnnounceID = uint32(1)
	ErrorID    = uint32(3)
)

type ConnectionResponse struct {
	Action        uint32
	TransactionID uint32
	ConnectionID  uint64
}

type AnnounceResponse struct {
	Action        uint32
	TransactionId uint32
	Interval      uint32
	Leechers      uint32
	Seeders       uint32

	Addresses []peerconnection.Peer
}

func sendConnectRequest(settings *settings.Settings) []byte {
	buffer := bytes.NewBuffer(make([]byte, 0, 16))

	binary.Write(buffer, binary.BigEndian, UDProtocol)
	binary.Write(buffer, binary.BigEndian, ConnectID)
	binary.Write(buffer, binary.BigEndian, settings.TransactionId)

	return buffer.Bytes()

}

func sendAnnounceRequest(t *TorrentFile, settings *settings.Settings) []byte {
	buffer := bytes.NewBuffer(make([]byte, 0, 98))

	binary.Write(buffer, binary.BigEndian, settings.ConnectionId)
	binary.Write(buffer, binary.BigEndian, AnnounceID)
	binary.Write(buffer, binary.BigEndian, settings.TransactionId)
	binary.Write(buffer, binary.BigEndian, t.InfoHash)

	binary.Write(buffer, binary.BigEndian, settings.PeerId)
	binary.Write(buffer, binary.BigEndian, uint64(0))
	binary.Write(buffer, binary.BigEndian, t.Length)
	binary.Write(buffer, binary.BigEndian, uint64(0))

	return buffer.Bytes()
}

func parseConnectionResponse(response []byte) ConnectionResponse {
	return ConnectionResponse{
		Action:        binary.BigEndian.Uint32(response[:4]),
		TransactionID: binary.BigEndian.Uint32(response[4:8]),
		ConnectionID:  binary.BigEndian.Uint64(response[8:16]),
	}
}

func (resp ConnectionResponse) ValidateConnectResponse(settings *settings.Settings) error {
	if resp.Action == ConnectID && resp.TransactionID == settings.TransactionId {
		settings.ConnectionId = resp.ConnectionID
		return nil
	}
	return errors.New("failed to connect to tracker")
}

func parseAnnouceResponse(response []byte) AnnounceResponse {
	var addresses []peerconnection.Peer

	for i := 20; i+6 < len(response); i += 6 {

		newAddress := peerconnection.Peer{
			IP:   net.IP(joinBytes(response[i : i+4])),
			Port: binary.BigEndian.Uint16(response[i+4 : i+6]),
		}

		addresses = append(addresses, newAddress)
	}

	return AnnounceResponse{
		Action:        binary.BigEndian.Uint32(response[:4]),
		TransactionId: binary.BigEndian.Uint32(response[4:8]),
		Seeders:       binary.BigEndian.Uint32(response[8:12]),
		Leechers:      binary.BigEndian.Uint32(response[12:16]),
		// Seeders:       binary.BigEndian.Uint32(response[16:20]),
		Addresses: addresses,
	}
}

func joinBytes(target []byte) string {
	return fmt.Sprintf("%d.%d.%d.%d", target[0], target[1], target[2], target[3])
}
