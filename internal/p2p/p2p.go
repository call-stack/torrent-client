package p2p

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/call-stack/torrent-client/internal/client"
	"github.com/call-stack/torrent-client/internal/message"
	"github.com/call-stack/torrent-client/internal/peers"
)

// MaxBlockSize is the largest number of bytes a request can ask for
const MaxBlockSize = 16384

// MaxBacklog is the number of unfulfilled requests a client can have in its pipeline
const MaxBacklog = 5

type Torrent struct {
	Peers       []peers.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHash   [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type PieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type PieceResult struct {
	index int
	buf   []byte
}

type pieceProgress struct {
	index      int
	client     *client.Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

func (t *Torrent) calculateBoundForPiece(index int) (int, int) {
	begin := index * t.PieceLength
	end := begin + t.PieceLength

	if end > t.Length {
		end = t.Length
	}
	return begin, end
}

func (t *Torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBoundForPiece(index)
	return end - begin
}

func (t *Torrent) startDownloadingWorker(peer peers.Peer, workerQueue chan *PieceWork, workerResult chan *PieceResult) {
	c, err := client.New(peer, t.InfoHash, t.PeerID)
	if err != nil {
		log.Printf("Could not handshake with %s. Disconnecting\n", peer.IP)
		return
	}
	defer c.Conn.Close()
	log.Printf("Completed handshake with %s\n", peer.IP)

	c.SendUnChoke()
	c.SendInterested()

	for pw := range workerQueue {
		if !c.Bitfield.HasPiece(pw.index) {
			workerQueue <- pw
			continue
		}

		buf, err := attemptDownloadPiece(c, pw)
		if err != nil {
			log.Println("Exit", err)
			workerQueue <- pw
			continue
		}

		err = checkIntegrity(pw, buf)
		if err != nil {
			log.Printf("Piece #%d failed integrity check\n", pw.index)
			workerQueue <- pw // Put piece back on the queue
			continue
		}

		c.SendHave(pw.index)
		workerResult <- &PieceResult{pw.index, buf}

	}

}

func (t *Torrent) Download() ([]byte, error) {
	log.Println("Starting the download....")
	workerQueue := make(chan *PieceWork, len(t.PieceHash))
	workerResult := make(chan *PieceResult)
	for index, hash := range t.PieceHash {
		length := t.calculatePieceSize(index)
		workerQueue <- &PieceWork{index, hash, length}
	}

	for _, peer := range t.Peers {
		go t.startDownloadingWorker(peer, workerQueue, workerResult)
	}

	buf := make([]byte, t.Length)
	donePieces := 0
	for donePieces < len(t.PieceHash) {
		res := <-workerResult
		begin, end := t.calculateBoundForPiece(res.index)
		copy(buf[begin:end], res.buf)
		donePieces++
		percent := (float64(donePieces) / float64(len(t.PieceHash))) * 100
		numWorkers := runtime.NumGoroutine() - 1
		log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, res.index, numWorkers)
	}
	close(workerQueue)
	return nil, nil
}

func attemptDownloadPiece(c *client.Client, pw *PieceWork) ([]byte, error) {
	state := pieceProgress{
		index:  pw.index,
		client: c,
		buf:    make([]byte, pw.length),
	}

	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{})

	for state.downloaded < pw.length {
		if !state.client.Choked {
			for state.backlog < MaxBacklog && state.requested < pw.length {
				blockSize := MaxBlockSize
				if pw.length-state.requested < blockSize {
					blockSize = pw.length - state.requested
				}

				err := c.SendRequest(pw.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}
				state.backlog++
				state.requested += blockSize

			}
		}

		err := state.readMessage()
		if err != nil {
			return nil, err
		}
	}

	return state.buf, nil
}

func (state *pieceProgress) readMessage() error {
	msg, err := state.client.Read()
	if err != nil {
		return err
	}

	if msg == nil {
		return nil
	}

	switch msg.ID {
	case message.MsgUnchoke:
		state.client.Choked = false
	case message.MsgChoke:
		state.client.Choked = true
	case message.MsgHave:
		idx, err := message.ParseHave(msg)
		if err != nil {
			return err
		}

		state.client.Bitfield.SetPiece(idx)
	case message.MsgPiece:
		n, err := message.ParsePiece(state.index, state.buf, msg)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlog--
	}

	return nil
}

func checkIntegrity(pw *PieceWork, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("Index %d failed integrity check", pw.index)
	}
	return nil

}
