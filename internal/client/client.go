package client

import (
	"bytes"
	"fmt"
	"github.com/kalpitpant/torrent-client/internal/bitfield"
	"github.com/kalpitpant/torrent-client/internal/handshake"
	"github.com/kalpitpant/torrent-client/internal/peers"
	"net"
	"time"
)

type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield bitfield.Bitfield
	peer     peers.Peer
	infoHash [20]byte
	peerID   [20]byte
}

func completeHandshake(conn net.Conn, infoHash, peerID [20]byte) (*handshake.Handshake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})

	fmt.Println("-----#4", peerID)
	req := handshake.New(infoHash, peerID)
	_, err := conn.Write(req.Serialize())
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 1)
	conn.Read(buf)



	res, err := handshake.Read(conn)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(res.InfoHash[:], infoHash[:]) {
		return nil, fmt.Errorf("expected infohash %x but got %x", res.InfoHash, infoHash)
	}
	return res, nil

}

func New(peer peers.Peer, infoHash, peerID [20]byte) (*Client, error) {
	fmt.Println("-----#3", peerID)
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}

	_, err = completeHandshake(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return nil, nil

}
