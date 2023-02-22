package seeder

import (
	"encoding/binary"
	peerconnect "matot/peerConnect"
	"matot/torrent"
)

type Request struct {
	SpMsg      peerconnect.SpecialMsg
	BlockBegin int
	BlockSize  int
}

func RecieveReqMsg(m *peerconnect.Message, t *torrent.TorrentFile) *Request {

	length := t.PieceLength * uint64(len(t.Pieces))

	index := binary.BigEndian.Uint32(m.Payload[0:4])
	requested := binary.BigEndian.Uint32(m.Payload[4:8])
	blockLength := binary.BigEndian.Uint32(m.Payload[8:12])

	begin := index*uint32(t.PieceLength) + requested
	end := begin + blockLength

	if end > uint32(length) {
		end = uint32(length)
		blockLength = end - begin
	}

	req := Request{}

	req.SpMsg.Index = index
	req.SpMsg.Begin = requested
	req.SpMsg.Length = end

	req.BlockBegin = int(begin)
	req.BlockSize = int(blockLength)

	return &req

}

func SendPiece(request *Request, piece []byte) *peerconnect.Message {
	payload := make([]byte, 8+len(piece))
	binary.BigEndian.PutUint32(payload[0:4], uint32(request.SpMsg.Index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(request.BlockBegin))
	for i := 0; i < len(piece); i++ {
		offset := 8
		payload[offset+i] = piece[i]
	}
	return &peerconnect.Message{ID: peerconnect.Piece, Payload: payload}
}
