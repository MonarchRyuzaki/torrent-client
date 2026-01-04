package main

import (
	"bytes"
	"encoding/binary"
	"io"
)

type messageID uint8

const (
	MsgChoke         messageID = 0
	MsgUnchoke       messageID = 1
	MsgInterested    messageID = 2
	MsgNotInterested messageID = 3
	MsgHave          messageID = 4
	MsgBitfield      messageID = 5
	MsgRequest       messageID = 6
	MsgPiece         messageID = 7
	MsgCancel        messageID = 8
)

type Message struct {
	ID      messageID
	Payload []byte
}

// <length 4 bytes><ID 1 byte><Payload>
func ReadMessage(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf)

	if length == 0 {
		return nil, nil
	}

	messageBuf := make([]byte, length)

	_, e := io.ReadFull(r, messageBuf)

	if e != nil {
		return nil, e
	}

	return &Message{
		ID:      messageID(messageBuf[0]),
		Payload: messageBuf[1:],
	}, nil
}

func SendMessage(w io.Writer, id byte, payload []byte) error {
	buf := make([]byte, 5)

	binary.BigEndian.PutUint32(buf[0:4], uint32(1))
	buf[4] = id
	newBuf := new(bytes.Buffer)

	newBuf.Write(buf)
	if payload != nil {
		newBuf.Write(payload)
	}
	
	_, err := w.Write(newBuf.Bytes())

	return err
}
