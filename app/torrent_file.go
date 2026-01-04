package main

import (
	"encoding/hex"
	"fmt"
)

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	Length      int
	Name        string
	PieceLength int
	PieceHashes [][20]byte
}

func pieceHash(piece [][20]byte) []string {
	hexHashes := make([]string, len(piece))
	for i, e := range piece {
		hexHashes[i] = hex.EncodeToString(e[:])
	}
	return hexHashes
}

func (tf *TorrentFile) String() string {
	return fmt.Sprintf("Tracker URL: %s\nLength: %d\nName: %s\nInfoHash: %s\nPiece Length: %d\nPiece Hashes: %v",
		tf.Announce,
		tf.Length,
		tf.Name,
		hex.EncodeToString(tf.InfoHash[:]),
		tf.PieceLength,
		pieceHash(tf.PieceHashes))
}
