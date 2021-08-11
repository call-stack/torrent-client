package torrent_file

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/call-stack/torrent-client/internal/peers"
	"github.com/jackpal/bencode-go"
)

type TrackerResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func (tf *TorrentFile) buildTrackerURL(peerID [20]byte, port int) (string, error) {
	base, err := url.Parse(tf.Announce)
	if err != nil {
		return "", nil
	}

	params := url.Values{
		"info_hash":  []string{string(tf.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(tf.Length)},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil

}

func (tf *TorrentFile) requestPeers(peerID [20]byte, port int) ([]peers.Peer, error) {
	//send the request to get the peers.
	URL, err := tf.buildTrackerURL(peerID, port)
	if err != nil {
		return nil, err
	}

	c := &http.Client{Timeout: 15 * time.Second}
	resp, err := c.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	trackerResp := TrackerResponse{}
	err = bencode.Unmarshal(resp.Body, &trackerResp)
	if err != nil {
		return nil, err
	}

	return peers.UnmarshalPeer([]byte(trackerResp.Peers))
}
