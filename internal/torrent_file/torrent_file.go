package torrent_file

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/jackpal/bencode-go"
	"github.com/kalpitpant/torrent-client/internal/p2p"
	"math/rand"
	"os"
)

type bencodedInfo struct {
	Pieces       string `bencode:"pieces"`
	PiecesLength int    `bencode:"piece length"`
	Length       int    `bencode:"length"`
	Name         string `bencode:"name"`
}

type bencodedTorrent struct {
	Announce string       `bencode:"announce"`
	Info     bencodedInfo `bencode:"info"`
}

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHash   [][20]byte
	PieceLength int
	Length      int
	Name        string
}

func Open(Path string) (TorrentFile, error) {
	file, err := os.Open(Path)
	if err != nil {
		return TorrentFile{}, err
	}
	bto := bencodedTorrent{}
	err = bencode.Unmarshal(file, &bto)
	if err != nil {
		return TorrentFile{}, err
	}

	return bto.toTorrentFile()
}

func (bi *bencodedInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *bi)
	if err != nil {
		return [20]byte{}, err
	}

	h := sha1.Sum(buf.Bytes())

	return h, nil
}

func (bi *bencodedInfo) splitPiecesHash() ([][20]byte, error) {
	hashLen := 20
	buf := []byte(bi.Pieces)
	if len(buf)%hashLen != 0 {
		return nil, fmt.Errorf("received malformed pieces of length")
	}

	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil

}

func (bto *bencodedTorrent) toTorrentFile() (TorrentFile, error) {

	infoHash, err := bto.Info.hash()
	if err != nil {
		return TorrentFile{}, err
	}

	piecesHashes, err := bto.Info.splitPiecesHash()
	if err != nil {
		return TorrentFile{}, err
	}

	torrentFile := TorrentFile{
		Announce:    bto.Announce,
		Length:      bto.Info.Length,
		PieceLength: bto.Info.PiecesLength,
		Name:        bto.Info.Name,
		PieceHash:   piecesHashes,
		InfoHash:    infoHash,
	}

	return torrentFile, nil
}

func (tf *TorrentFile) DownloadToFile() error {
	//lets create your peers Id
	var peerID [20]byte
	_, err := rand.Read(peerID[:])

	if err != nil {
		return err
	}
	fmt.Println("-----#1", peerID)
	peers, err := tf.requestPeers(peerID, PORT)
	if err != nil {
		return err
	}

	torrent := &p2p.Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    tf.InfoHash,
		PieceHash:   tf.PieceHash,
		PieceLength: tf.PieceLength,
		Length:      tf.Length,
		Name:        tf.Name,
	}

	_, err = torrent.Download()
	if err!=nil{
		return err
	}


	return nil
}
