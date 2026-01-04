package main

import "bytes"

type Handshake struct {
	handshakeMsg []byte
	InfoHash     [20]byte
	PeerID       [20]byte
}

var handshake = Handshake{}

func (hs *Handshake) prepareHandshakeMessage() {
	buf := new(bytes.Buffer)

	buf.WriteByte(19)

	buf.WriteString("BitTorrent protocol")

	buf.Write(make([]byte, 8))

	buf.Write(hs.InfoHash[:])

	buf.Write(hs.PeerID[:])

	hs.handshakeMsg = buf.Bytes()
}
