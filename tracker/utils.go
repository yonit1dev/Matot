package tracker

import (
	"encoding/binary"
	"log"
	"net"
	"strconv"
)

// ParsePeerAddress parses peer IP addresses and ports from a buffer
func ParsePeerAddress(peers []byte) ([]Peer, error) {
	const peerAddressSize = 6 // 4 for IP, 2 for port
	totalPeers := len(peers) / peerAddressSize

	if len(peers)%peerAddressSize != 0 {
		log.Fatal("peer addresses corrupt")
	}

	peerAdd := make([]Peer, totalPeers)
	for i := 0; i < totalPeers; i++ {
		offset := i * peerAddressSize
		peerAdd[i].IP = net.IP(peers[offset : offset+4])
		peerAdd[i].Port = binary.BigEndian.Uint16([]byte(peers[offset+4 : offset+6]))
	}
	return peerAdd, nil
}

// function that joins parsed peer ip addresses and ports
func (peer Peer) String() string {
	return net.JoinHostPort(peer.IP.String(), strconv.Itoa(int(peer.Port)))
}
