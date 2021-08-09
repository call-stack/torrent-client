package peers

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

func UnmarshalPeer(peerbin []byte) ([]Peer, error) {
	peerSize := 6 // As compact is set to 1 while making the request.

	if len(peerbin)%peerSize != 0 {
		return nil, fmt.Errorf("received malformed peers")
	}

	numberOfPeers := len(peerbin) / peerSize
	peers := make([]Peer, numberOfPeers)
	for i := 0; i < numberOfPeers; i++ {
		offset := i * peerSize
		peers[i].IP = net.IP(peerbin[offset : offset+4])
		peers[i].Port = binary.BigEndian.Uint16(peerbin[offset+4 : offset+6])

	}

	return peers, nil
}

func (p Peer) String() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}

