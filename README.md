# Matot
Matot is a minimal BitTorrent client built using Go.

#### Features
- Supports TCP and UDP trackers.
- Leeching and Seeding.
- Resuming partial downloads.

### Explanation about the Bittorrent Protocol
Bittorrent protocol is a peer to peer file sharing protocol. It works over transport layer of internet protocol networks.
  - A bittorrent client is a software that runs on a machine that is able to connect to other peers to send (seeding) or recieve (leeching) files.
  - A shared file is broken down into pieces and peers only share the pieces they have by a series of exchanging messages.
  - For more information read the [formal specification of Bittorrent](https://www.bittorrent.org/beps/bep_0003.html). A more general unofficial overview can also be found [here](https://wiki.theory.org/BitTorrentSpecification).

The process of building a client can be broken down into three major stages:
  1. Parsing the meta-info(torrent file)
    - Torrent file contains all necessary information about a file and it's location.
    - It's in bencode format (a type of formatting, think something like JSON).
    - Parse the torrent for use to contact the tracker.

  2. Contacting a tracker (central) server to get list of peers.
    - A tracker is a central server that peers contact inorder to get list of peers that own the file they're interested in.
    - A tracker can be a UDP or TCP server.
    - The tracker expects information about the peer, the info hash(hash of the info section of the torrent) and other information.
    - The tracker returns a list of peers to the requesting client.

  3. Connecting to peers and sharing data.
    - Peers share information over TCP layer.
    - They exhange a series of messages regarding the pieces they have.
    - In this specific implementation, each connection with a peer is a go routine.
    - Go routines communicate over two channels, one to handle the current piece being downloaded, and another to communicate the results of downloads.
    - Each piece is written to a local file, to enable resumption of downloads.

The above steps describe a leeching client. To be able to seed (upload) files, the reverse process happens.
  - Connections to requesting peers are go routines.
  - The client sends a bitfield message to peer connection.
  - Then, it parses a request message, and looks for the piece from the local file.
  - It sends the piece over the connection.


#### How to use
- Clone this repository.
- Run `go mody tidy` to fetch necessary packages.
- Change the torrent file directory in the 'main.go' file.
- Build and run the module:
  ```
  go build
  ./matot
  ```
