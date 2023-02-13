package peerconnect

type BitFieldType []byte

func (b BitFieldType) PieceExist(index int) bool {
	bytePos := index / 8
	bitPos := index % 8

	/*
	 * check if request piece index is within bitfield bounds
	 * following if check return false if the index is within the bounds
	 */

	if bytePos >= (len(b)) || bytePos < 0 {
		return false
	}

	return b[bytePos]>>uint(7-bitPos)&1 != 0
}

func (b BitFieldType) ChangeBit(index int) {
	bytePos := index / 8
	bitPos := index % 8

	if bytePos < 0 || bytePos >= (len(b)) {
		return
	}
	b[bytePos] |= 1 << uint(7-bitPos)
}
