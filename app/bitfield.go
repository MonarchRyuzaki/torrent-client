package main

type Bitfield []byte

func (b Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	byteOffset := index % 8

	if byteIndex < 0 || byteIndex >= len(b) {
		return false
	}

	return ((b[byteIndex] >> (7 - byteOffset)) & 1) != 0
}

func (b Bitfield) SetPiece(index int) {
	byteIndex := index / 8
	byteOffset := index % 8

	if byteIndex < 0 || byteIndex >= len(b) {
		return
	}

	b[byteIndex] |= (1 << (7 - byteOffset))
}
