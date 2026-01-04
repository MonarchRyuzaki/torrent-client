package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
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

func (p *Peer) Download(tf *TorrentFile, piece_index int) {
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

	msg, err := ReadMessage(conn)
	if err != nil {
		fmt.Println("Cant Read Bitfield Message", err)
		return
	}

	fmt.Printf("Received Message: ID %d (Bitfield)\n", msg.ID)

	SendMessage(conn, 2, nil)

	for {
		msg, err := ReadMessage(conn)
		if err != nil {
			return
		}

		if msg == nil {
			continue
		}

		if msg.ID == 1 {
			fmt.Println("Received: Unchoke! We can download now.")
			break
		}
	}

	var piece_length int
	if piece_index == len(tf.PieceHashes)-1 {
		piece_length = tf.Length - (len(tf.PieceHashes)-1)*tf.PieceLength
	} else {
		piece_length = tf.PieceLength
	}

	buf := make([]byte, piece_length)

	const BlockSize = 16 * 1024
	numBlocks := (piece_length + BlockSize - 1) / BlockSize

	for i := range numBlocks {
		blockPayload := make([]byte, 12)

		begin := i * BlockSize
		blockLength := BlockSize
		if begin+blockLength > piece_length {
			blockLength = piece_length - begin
		}

		binary.BigEndian.PutUint32(blockPayload[0:4], uint32(piece_index))
		binary.BigEndian.PutUint32(blockPayload[4:8], uint32(begin))
		binary.BigEndian.PutUint32(blockPayload[7:12], uint32(blockLength))
		fmt.Printf("Requesting: Index: %d, Begin: %d, Length: %d\n", piece_index, begin, blockLength)
		err := SendMessage(conn, 6, blockPayload)
		if err != nil {
			fmt.Println("Failed to send request:", err)
			return
		}
	}

	for i := range numBlocks {
		var _ = i
		msg, err := ReadMessage(conn)
		if err != nil {
			fmt.Println(err)
			return
		}
		// block_index := binary.BigEndian.Uint32(msg.Payload[0:4])
		block_begin := binary.BigEndian.Uint32(msg.Payload[4:8])
		block_length := BlockSize
		if int(block_begin)+block_length > piece_length {
			block_length = piece_length - int(block_begin)
		}
		block_data := msg.Payload[8:]

		copy(buf[block_begin:block_begin+uint32(block_length)], block_data)
	}

	ok, err := verifyIntegrityOfPiece(buf, tf.PieceHashes[piece_index])
	if err != nil {
		fmt.Println(err)
	}
	if !ok {
		fmt.Println("Piece is not valid")
		return
	}
	file, err := os.Create(fmt.Sprintf("tmp-%s-%d", tf.Name, piece_index))
	if err != nil {
		log.Fatalf("Cant Create file")
	}
	defer file.Close()
	file.Write(buf)
}

func verifyIntegrityOfPiece(buf []byte, b [20]byte) (bool, error) {
	h := sha1.New()

	_, err := h.Write(buf)
	if err != nil {
		log.Fatalf("Error writing to hash: %v", err)
		return false, err
	}
	hashInBytes := h.Sum(nil)
	if bytes.Equal(hashInBytes, b[:]) {
		return true, nil
	}
	return false, nil
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
