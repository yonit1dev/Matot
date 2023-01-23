package torrenttracker

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"time"
	"torrentClient/settings"
)

type Client struct {
	conn net.Conn
}

func (client *Client) Close() error {
	return client.conn.Close()
}

func CreateClient(t *TorrentFile) *Client {
	url, err := url.Parse(t.Announce)
	if err != nil {
		log.Fatal("unable to parse tracker url")
	}

	conn, err := net.Dial("udp", url.Host)
	if err != nil {
		log.Fatal("unable to connect to the tracker")
	}

	conn.SetReadDeadline(time.Now().Add(45 * time.Second))
	defer conn.SetReadDeadline(time.Time{})

	return &Client{conn: conn}
}

func (client *Client) ConnectTracker(t *TorrentFile, settings *settings.Settings) ConnectionResponse {

	var trackerResponse = make([]byte, 16)

	if _, err := client.conn.Write(sendConnectRequest(settings)); err != nil {

		log.Fatal("Unable to send message to the tracker")
	}

	if _, err := client.conn.Read(trackerResponse); err != nil {
		log.Fatal("unable to read from the tracker")
	}

	return parseConnectionResponse(trackerResponse)
}

func (client *Client) AnnounceTracker(t *TorrentFile, settings *settings.Settings) AnnounceResponse {
	if _, err := client.conn.Write(sendAnnounceRequest(t, settings)); err != nil {
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
