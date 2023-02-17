package tracker

import (
	"fmt"
	"log"
	"matot/config"
	"matot/torrent"
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
		log.Fatalf("Response status code error! Expected: %d, Got: %d", 200, response.StatusCode)
	}

	trackerResponse := TrackerResponse{}
	err = bencode.Unmarshal(response.Body, &trackerResponse)
	if err != nil {
		log.Fatal(err)
	}

	if trackerResponse.FailureReason != "" {
		log.Fatalf(trackerResponse.FailureReason)
	}

	// Display warning message
	if trackerResponse.WarningMessage != "" {
		fmt.Println(trackerResponse.WarningMessage)
	}

	peers, err := ParsePeerAddress([]byte(trackerResponse.Peers))
	if err != nil {
		return 0, nil
	}

	return trackerResponse.Interval, peers
}
