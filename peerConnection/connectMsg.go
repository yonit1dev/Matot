package peerconnection

func (c *PeerClient) SendUnchokeMsg() error {
	message := Message{
		ID: UnChoke,
	}
	_, err := c.Conn.Write(message.BufferMessage())
	return err

}
func (c *PeerClient) SendInteresetedMsg() error {
	message := Message{
		ID: Intereseted,
	}
	_, err := c.Conn.Write(message.BufferMessage())
	return err

}

func (c *PeerClient) SendNotInteresetedMsg() error {
	message := Message{
		ID: NotInterested,
	}
	_, err := c.Conn.Write(message.BufferMessage())
	return err

}

func (c *PeerClient) SendRequestMsg(index, begin, length int) error {
	message := RequestMsg(index, begin, length)
	_, err := c.Conn.Write(message.BufferMessage())
	return err

}

func (c *PeerClient) SendHaveMsg(index int) error {
	message := HaveMsg(index)
	_, err := c.Conn.Write(message.BufferMessage())
	return err
}
