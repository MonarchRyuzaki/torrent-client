package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"strings"

	"github.com/jackpal/bencode-go"
)

type BencodeInfo struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
}

type BencodeFile struct {
	Announce string      `bencode:"announce"`
	Info     BencodeInfo `bencode:"info"`
}

func (bf *BencodeFile) toTorrentFile() (*TorrentFile, error) {
	tf := TorrentFile{}
	tf.Announce = bf.Announce
	tf.Name = bf.Info.Name
	tf.Length = bf.Info.Length
	tf.PieceLength = bf.Info.PieceLength

	pieceHashes, err := splitPieceHashes(bf.Info.Pieces)
	if err != nil {
		return nil, err
	}
	tf.PieceHashes = pieceHashes

	infoHash, err := calculateInfoHash(bf.Info)
	if err != nil {
		return nil, err
	}
	tf.InfoHash = infoHash

	return &tf, nil
}

func splitPieceHashes(piecesStr string) ([][20]byte, error) {
	const SHA1Size = sha1.Size
	if len(piecesStr)%SHA1Size != 0 {
		return nil, fmt.Errorf("received malformed pieces of length %d", len(piecesStr))
	}
	numHashes := len(piecesStr) / SHA1Size
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		start := i * SHA1Size
		end := start + SHA1Size
		chunk := piecesStr[start:end]
		copy(hashes[i][:], chunk)
	}

	return hashes, nil
}

func calculateInfoHash(info BencodeInfo) ([20]byte, error) {
	var infoDictBencoded strings.Builder
	e := bencode.Marshal(&infoDictBencoded, info)
	if e != nil {
		return [20]byte{}, e
	}
	infoDictBencodedString := infoDictBencoded.String()
	h := sha1.New()

	_, err := h.Write([]byte(infoDictBencodedString))
	if err != nil {
		log.Fatalf("Error writing to hash: %v", err)
	}
	hashInBytes := h.Sum(nil)
	return [20]byte(hashInBytes), nil
}
