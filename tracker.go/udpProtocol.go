package tracker

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"matot/config"
	"matot/torrent"
	"net"
)

const UDProtocol = uint64(0x41727101980)

const (
	ConnectID  = uint32(0)
	AnnounceID = uint32(1)
	ScrapaeID  = uint(2)
	ErrorID    = uint32(3)
)

type ConnectionResponse struct {
	Action        uint32
	TransactionID uint32
	ConnectionID  uint64
}

type AnnounceResponse struct {
	Action        uint32
	TransactionID uint32
	Interval      uint32
	Leechers      uint32
	Seeders       uint32

	Addresses []Peer
}

func sendConnectRequest(config *config.Config) []byte {
	buffer := bytes.NewBuffer(make([]byte, 0, 16))

	binary.Write(buffer, binary.BigEndian, UDProtocol)
	binary.Write(buffer, binary.BigEndian, ConnectID)
	binary.Write(buffer, binary.BigEndian, config.TransactionID)

	return buffer.Bytes()

}

func sendAnnounceRequest(t *torrent.TorrentFile, config *config.Config) []byte {
	buffer := bytes.NewBuffer(make([]byte, 0, 98))

	// connectionId
	binary.Write(buffer, binary.BigEndian, config.ConnectionID)
	// action
	binary.Write(buffer, binary.BigEndian, AnnounceID)
	// transactionId
	binary.Write(buffer, binary.BigEndian, config.TransactionID)
	// info_hash
	binary.Write(buffer, binary.BigEndian, t.InfoHash)

	// peer_id
	binary.Write(buffer, binary.BigEndian, config.PeerID)
	// downloaded
	binary.Write(buffer, binary.BigEndian, uint64(0))
	// left
	binary.Write(buffer, binary.BigEndian, t.Length)
	// uploaded
	binary.Write(buffer, binary.BigEndian, uint64(0))
	//event
	binary.Write(buffer, binary.BigEndian, uint32(0))
	// ip address
	binary.Write(buffer, binary.BigEndian, uint32(0))
	// key
	binary.Write(buffer, binary.BigEndian, uint32(2425))
	// num_want
	binary.Write(buffer, binary.BigEndian, int32(-1))
	// port
	binary.Write(buffer, binary.BigEndian, config.Port)

	return buffer.Bytes()
}

func parseConnectionResponse(response []byte) ConnectionResponse {
	return ConnectionResponse{
		Action:        binary.BigEndian.Uint32(response[:4]),
		TransactionID: binary.BigEndian.Uint32(response[4:8]),
		ConnectionID:  binary.BigEndian.Uint64(response[8:16]),
	}
}

func (resp ConnectionResponse) ValidateConnectResponse(config *config.Config) error {
	if resp.Action == ConnectID && resp.TransactionID == config.TransactionID {
		config.ConnectionID = uint32(resp.ConnectionID)
		return nil
	}
	return errors.New("failed to connect to tracker. connect error")
}

func parseAnnouceResponse(response []byte) AnnounceResponse {
	var addresses []Peer

	for i := 20; i+6 < len(response); i += 6 {

		newAddress := Peer{
			IP:   net.IP(joinBytes(response[i : i+4])),
			Port: binary.BigEndian.Uint16(response[i+4 : i+6]),
		}

		addresses = append(addresses, newAddress)
	}

	return AnnounceResponse{
		Action:        binary.BigEndian.Uint32(response[:4]),
		TransactionID: binary.BigEndian.Uint32(response[4:8]),
		Interval:      binary.BigEndian.Uint32(response[8:12]),
		Leechers:      binary.BigEndian.Uint32(response[12:16]),
		Seeders:       binary.BigEndian.Uint32(response[16:20]),
		Addresses:     addresses,
	}
}

func joinBytes(target []byte) string {
	return fmt.Sprintf("%d.%d.%d.%d", target[0], target[1], target[2], target[3])
}
