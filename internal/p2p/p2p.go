package p2p

import (
	"fmt"
	"github.com/kalpitpant/torrent-client/internal/client"
	"github.com/kalpitpant/torrent-client/internal/peers"
	"log"
)

type Torrent struct {
	Peers       []peers.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHash   [][20]byte
	PieceLength int
	Length      int
	Name        string
}

func (t *Torrent) startDownloadingWorker(peer peers.Peer) {
	fmt.Println("-----#2", t.PeerID)
	c, err := client.New(peer, t.InfoHash, t.PeerID)
	if err != nil {
		log.Printf("Could not handshake with %s. Disconnecting\n", peer.IP)
		return
	}
	defer c.Conn.Close()
	log.Printf("Completed handshake with %s\n", peer.IP)
}

func (t *Torrent) Download() ([]byte, error) {
	log.Println("Starting the download....")

	for _, peer := range t.Peers {
		t.startDownloadingWorker(peer)
	}

	return nil, nil
}
