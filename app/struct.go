package main

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"log"
	"net"
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

type TrackerResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
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
	handshake.InfoHash = infoHash

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

func (tr *TrackerResponse) extractPeerIps() ([]Peer, error) {

	const PeerSize = 6
	if len(tr.Peers)%PeerSize != 0 {
		return nil, fmt.Errorf("malformed peers")
	}
	numPeers := len(tr.Peers) / PeerSize
	peers := make([]Peer, 0)

	for i := range numPeers {
		start := i * PeerSize
		end := start + PeerSize
		chunk := []byte(tr.Peers[start:end])
		newPeer := Peer{
			IP:     net.IPv4(chunk[0], chunk[1], chunk[2], chunk[3]),
			Port:   uint16(binary.BigEndian.Uint16(chunk[4:6])),
			Status: 2,
		}
		peers = append(peers, newPeer)
	}

	return peers, nil
}
