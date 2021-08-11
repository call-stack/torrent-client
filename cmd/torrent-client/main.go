package main

import (
	"fmt"

	"github.com/call-stack/torrent-client/internal/torrent_file"
)

func main() {

	tf, err := torrent_file.Open("filename.torrent")
	if err != nil {
		//then some issue here
		print("Error", err)
	}

	tf.DownloadToFile()

	fmt.Println("eer", err)
}
