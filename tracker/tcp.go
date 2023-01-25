package tracker

import (
	"go-torrent-client/peer"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jackpal/bencode-go"
)

type TrackerClient struct {
	torrent *TorrentFile
}

func NewTrackerClient(t *TorrentFile) *TrackerClient {
	return &TrackerClient{t}
}

type TrackerResponse struct {
	Interval int64
	Peers    string
}

func (tc *TrackerClient) prepareTrackerReq(peerId [20]byte, port uint16) (string, error) {
	baseUrl, err := url.Parse(tc.torrent.Announce)
	if err != nil {
		log.Fatalf("failed to parse tracker url. Error: %s", err)
	}

	params := url.Values{
		"info_hash":  []string{string(tc.torrent.InfoHash[:])},
		"peer_id":    []string{string(peerId[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(int(tc.torrent.Length))},
	}

	baseUrl.RawQuery = params.Encode()
	return baseUrl.String(), nil
}
func (tc *TrackerClient) GetPeersTCP(peerId [20]byte, port uint16) ([]peer.Peer, error) {

	url, err := tc.prepareTrackerReq(peerId, port)
	if err != nil {
		log.Fatal(err.Error())
	}

	c := &http.Client{Timeout: time.Second * 30}

	response, err := c.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	trackerResponse := TrackerResponse{}
	err = bencode.Unmarshal(response.Body, &trackerResponse)
	if err != nil {
		log.Fatal(err)
	}

	return peer.ParsePeerAddress([]byte(trackerResponse.Peers))
}
