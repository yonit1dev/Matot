package tracker

import (
	"encoding/binary"
	"goAssignment/config"
	"goAssignment/torrent"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jackpal/bencode-go"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

type TrackerClient struct {
	torrent *torrent.TorrentFile
}

func NewTrackerClient(t *torrent.TorrentFile) *TrackerClient {
	return &TrackerClient{t}
}

type TrackerResponse struct {
	Interval       int64  `bencode:"interval"`
	Peers          string `bencode:"peers"`
	FailureReason  string `bencode:"failure reason"`
	WarningMessage string `bencode:"warning message"`
}

func (tc *TrackerClient) prepareTrackerReq(config *config.Config) (string, error) {
	baseUrl, err := url.Parse(tc.torrent.Announce)
	if err != nil {
		log.Fatalf("failed to parse tracker url. Error: %s", err)
	}

	params := url.Values{
		"info_hash":  []string{string(tc.torrent.InfoHash[:])},
		"peer_id":    []string{string(config.PeerID[:])},
		"port":       []string{strconv.Itoa(int(config.Port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(int(tc.torrent.Length))},
	}

	baseUrl.RawQuery = params.Encode()
	return baseUrl.String(), nil
}

func (tc *TrackerClient) GetPeersTCP(config *config.Config) (int64, []Peer) {

	url, err := tc.prepareTrackerReq(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	c := &http.Client{Timeout: time.Second * 60}

	response, err := c.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Fatal("Response from tracker not ok!")
	}

	trackerResponse := TrackerResponse{}
	err = bencode.Unmarshal(response.Body, &trackerResponse)
	if err != nil {
		log.Fatal(err)
	}

	if trackerResponse.FailureReason != "" {
		log.Fatalf(trackerResponse.FailureReason)
	}

	peers, err := ParsePeerAddress([]byte(trackerResponse.Peers))
	if err != nil {
		return 0, nil
	}

	return trackerResponse.Interval, peers
}

// ParsePeerAddress parses peer IP addresses and ports from a buffer
func ParsePeerAddress(peers []byte) ([]Peer, error) {
	const peerAddressSize = 6 // 4 for IP, 2 for port
	totalPeers := len(peers) / peerAddressSize

	if len(peers)%peerAddressSize != 0 {
		log.Fatal("peer addresses corrupt")
	}

	peerAdd := make([]Peer, totalPeers)
	for i := 0; i < totalPeers; i++ {
		slice := i * peerAddressSize
		peerAdd[i].IP = net.IP(peers[slice : slice+4])
		peerAdd[i].Port = binary.BigEndian.Uint16([]byte(peers[slice+4 : slice+6]))
	}
	return peerAdd, nil
}

// function that joins parsed peer ip addresses and ports
func (peer Peer) String() string {
	return net.JoinHostPort(peer.IP.String(), strconv.Itoa(int(peer.Port)))
}
