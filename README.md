# Matot
Matot is a minimal BitTorrent client built using Go.

#### Features
- Supports TCP and UDP trackers.
- Leeching only.
- Resuming partial downloads.

### Explanation about the Bittorrent Protocol
Bittorrent protocol is a peer to peer file sharing protocol.

#### How to use
- Clone this repository.
- Run `go mody tidy` to fetch necessary packages.
- Change the torrent file directory in the 'main.go' file.
- Build and run the module:
  ```
  go build
  ./matot
  ```
