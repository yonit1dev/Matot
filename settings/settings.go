package settings

type Settings struct {
	TransactionId uint32
	ConnectionId  uint64
	Port          uint16
	PeerId        [20]byte
}
