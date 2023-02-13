package config

type Config struct {
	TransactionID uint32
	ConnectionID  uint32
	Port          uint16
	PeerID        [20]byte
}
