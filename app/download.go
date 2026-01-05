package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type DownloadState struct {
	pieceStatus     []int //0 -> Not downloaded, 1 -> Download in progress, 2 -> Download Done
	pieceLength     []int
	numOfPiecesLeft int
	mu              sync.Mutex
}

type PeerConnection struct {
	pConn []net.Conn
	mu    []sync.Mutex
}

func GetPeerConnection(peers []Peer, pc *PeerConnection) {
	var wg sync.WaitGroup
	for i := range peers {
		if peers[i].Status != 2 {
			continue
		}
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			conn, err := peers[index].EstablishHandshake()
			if err != nil {
				fmt.Printf("Cant Establish Handshake with %v : %v", peers[index], err)
				return
			}
			peers[index].Status = 0
			pc.pConn[index] = conn
		}(i)
	}

	wg.Wait()
}

func DownloadManager(tf *TorrentFile, peers []Peer) {
	pc := PeerConnection{
		pConn: make([]net.Conn, len(peers)),
		mu:    make([]sync.Mutex, len(peers)),
	}

	GetPeerConnection(peers, &pc)

	ds := DownloadState{
		pieceStatus:     make([]int, len(tf.PieceHashes)),
		pieceLength:     make([]int, len(tf.PieceHashes)),
		numOfPiecesLeft: len(tf.PieceHashes),
	}

	for {
		time.Sleep(1 * time.Second)

		if ds.numOfPiecesLeft == 0 {
			break
		}

		for piece_index, e := range ds.pieceStatus {
			if e != 0 {
				continue
			}

			assigned := false
			for peer_index := range peers {
				pc.mu[peer_index].Lock()

				if peers[peer_index].Status == 0 && peers[peer_index].Bitfield.HasPiece(piece_index) {
					peers[peer_index].Status = 1
					ds.pieceStatus[piece_index] = 1
					pc.mu[peer_index].Unlock()

					go func(peerIdx, pieceIdx int) {
						len, err := peers[peerIdx].DownloadPiece(pc.pConn[peerIdx], tf, pieceIdx)

						pc.mu[peerIdx].Lock()
						defer pc.mu[peerIdx].Unlock()

						if err != nil {
							fmt.Println("Download Failed:", err)
							pc.pConn[peerIdx].Close()
							peers[peerIdx].Status = 2
							ds.mu.Lock()
							ds.pieceStatus[pieceIdx] = 0
							ds.mu.Unlock()
						} else {
							ds.mu.Lock()
							peers[peerIdx].Status = 0
							ds.pieceLength[pieceIdx] = len
							ds.pieceStatus[pieceIdx] = 2
							ds.numOfPiecesLeft--
							fmt.Printf("Piece %d Done. Left: %d\n", pieceIdx, ds.numOfPiecesLeft)
							ds.mu.Unlock()
						}

					}(peer_index, piece_index)
					assigned = true
					break
				} else {
					pc.mu[peer_index].Unlock()
				}
			}
			if assigned {
				continue
			}
		}
	}

	for peer_index := range peers {
		if pc.pConn[peer_index] != nil {
			pc.pConn[peer_index].Close()
		}
	}

	mergePieces(tf, &ds)
}

func mergePieces(tf *TorrentFile, ds *DownloadState) {
	finalBuf := new(bytes.Buffer)

	for i := range ds.pieceStatus {
		currBuf := make([]byte, ds.pieceLength[i])
		file, err := os.Open(fmt.Sprintf("tmp-%v-%v", tf.Name, i))
		if err != nil {
			log.Fatalf("Cant Open file")
		}
		defer file.Close()
		_, e := io.ReadFull(file, currBuf)
		if e != nil {
			log.Fatalf("Cant Read File")
		}

		os.Remove(fmt.Sprintf("tmp-%v-%v", tf.Name, i))

		finalBuf.Write(currBuf)
	}

	file, err := os.Create(fmt.Sprintf("%s", tf.Name))
	if err != nil {
		log.Fatalf("Cant Create file")
	}
	defer file.Close()
	file.Write(finalBuf.Bytes())
}
