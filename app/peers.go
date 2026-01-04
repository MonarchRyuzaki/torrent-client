package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net"
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
	conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", p.IP, p.Port))
	if err != nil {
		fmt.Println("Error Connecting:", err)
		return
	}
	defer conn.Close()

	message := handshake.handshakeMsg

	_, e := conn.Write(message)
	if e != nil {
		fmt.Println("Error Writing:", e)
		return
	}

	// fmt.Printf("Sent to server: %v\n", message)

	reader := bufio.NewReader(conn)

	recvMsg := make([]byte, 100)
	n, err := reader.Read(recvMsg)
	if err != nil {
		fmt.Println("Cant Receive Handshake Message: ", err)
		return
	}

	p.deserializeHandshakeMessage(recvMsg[:n])

}

func (p *Peer) deserializeHandshakeMessage(b []byte) {
	p.PeerID = [20]byte(b[len(b)-20:])
}
