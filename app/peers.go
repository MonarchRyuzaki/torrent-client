package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"time"
)

type Peer struct {
	IP     net.IP
	Port   uint16
	PeerID [20]byte
}

func (p *Peer) String() string {
	return fmt.Sprintf("%v:%v PeerID:%v\n", p.IP, p.Port, hex.EncodeToString(p.PeerID[:]))
}

func (p *Peer) handleHandshake() {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%v:%v", p.IP, p.Port), 3*time.Second)
	if err != nil {
		fmt.Println("Error Connecting:", err)
		return
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	message := handshake.handshakeMsg

	_, e := conn.Write(message)
	if e != nil {
		fmt.Println("Error Writing:", e)
		return
	}

	// fmt.Printf("Sent to server: %v\n", message)

	recvMsg := make([]byte, 68)
	_, er := io.ReadFull(conn, recvMsg)
	if er != nil {
		fmt.Println("Cant Receive Handshake Message: ", er)
		return
	}

	p.deserializeHandshakeMessage(recvMsg)

}

func (p *Peer) deserializeHandshakeMessage(b []byte) {
	if len(b) < 68 {
		fmt.Println("Error: Handshake too short")
		return
	}

	if b[0] != 19 || string(b[1:20]) != "BitTorrent protocol" {
		fmt.Println("Error: Invalid protocol")
		return
	}

	if !bytes.Equal(b[28:48], handshake.InfoHash[:]) {
		fmt.Printf("Error: InfoHash mismatch. Expected %x, Got %x\n", handshake.InfoHash, b[28:48])
		return
	}

	copy(p.PeerID[:], b[48:])
}
