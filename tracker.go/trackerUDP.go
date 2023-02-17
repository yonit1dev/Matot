package tracker

import (
	"fmt"
	"log"
	"matot/config"
	"matot/torrent"
	"net"
	"net/url"
	"time"
)

type Client struct {
	conn net.Conn
}

func (client *Client) Close() error {
	return client.conn.Close()
}

func CreateClient(t *torrent.TorrentFile) *Client {
	url, err := url.Parse(t.Announce)
	if err != nil {
		log.Fatal("unable to parse tracker url")
	}

	conn, err := net.Dial("udp", url.Host)
	if err != nil {
		log.Fatal("unable to connect to the tracker")
	}

	conn.SetReadDeadline(time.Now().Add(45 * time.Second))
	// defer conn.SetReadDeadline(time.Time{})

	return &Client{conn: conn}
}

func (client *Client) ConnectTracker(config *config.Config) ConnectionResponse {
	connectReq := sendConnectRequest(config)

	_, err := client.conn.Write(connectReq)
	if err != nil {
		log.Fatal("Unable to send message to the tracker")
	}

	var trackerResponse = make([]byte, 16)

	n, err := client.conn.Read(trackerResponse)
	if n < 16 && err != nil {
		log.Fatal("unable to read from the tracker")
	}

	return parseConnectionResponse(trackerResponse)
}

func (client *Client) AnnounceTracker(t *torrent.TorrentFile, config *config.Config) AnnounceResponse {
	if _, err := client.conn.Write(sendAnnounceRequest(t, config)); err != nil {
		log.Fatal("Unable to send announce request to the tracker")
	}

	response := make([]byte, 1024)
	n, err := client.conn.Read(response)
	if err != nil {
		log.Fatal(response, "\n", "Unable to read announce response from the tracker: ", err)
	}

	fmt.Println(n)

	return parseAnnouceResponse(response[:n])
}
