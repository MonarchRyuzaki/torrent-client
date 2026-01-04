package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jackpal/bencode-go"
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

func (tf *TorrentFile) buildTrackerUrl(peerId [20]byte) (string, error) {
	baseUrl, err := url.Parse(tf.Announce)
	if err != nil {
		return "", err
	}
	params := url.Values{
		"info_hash":  []string{string(tf.InfoHash[:])},
		"peer_id":    []string{string(peerId[:])},
		"port":       []string{"6881"},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(tf.Length)},
	}
	baseUrl.RawQuery = params.Encode()
	return baseUrl.String(), nil
}

func generateUniqueString(length int) ([]byte, error) {

	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

func (tf *TorrentFile) getPeers() ([]Peer, error) {
	peerId, err := generateUniqueString(20)
	if err != nil {
		return nil, err
	}
	trackerUrl, err := tf.buildTrackerUrl([20]byte(peerId))
	if err != nil {
		return nil, err
	}
	resp, err := http.Get(trackerUrl)
	if err != nil {
		log.Fatalf("Error making GET request: %v", err)
	}
	defer resp.Body.Close()
	var peers []Peer
	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Error reading response body: %v", err)
			return nil, err
		}
		trackerResp := TrackerResponse{}
		e := bencode.Unmarshal(strings.NewReader(string(body)), &trackerResp)
		if e != nil {
			return nil, e
		}
		peers, e = trackerResp.extractPeerIps()
		if e != nil {
			return nil, e
		}
	} else {
		return nil, fmt.Errorf("Request failed with status code: %d\n", resp.StatusCode)
	}
	return peers, nil
}
