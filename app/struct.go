package main

type BencodeInfo struct {
	Length int `bencode:"length"`
	Name string `bencode:"name"`
	PieceLength int `bencode:"piece length"`
	Pieces string `bencode:"pieces"`
}

type BencodeFile struct {
	Announce string `bencode:"announce"`
	Info BencodeInfo `bencode:"info"`
}
