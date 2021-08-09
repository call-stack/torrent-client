package main

import (
	"fmt"
	"github.com/kalpitpant/torrent-client/internal/torrent_file"
)

func main() {
	fmt.Println("Random values------")
	tf, err := torrent_file.Open("/home/kalpit/Downloads/debian-10.10.0-amd64-netinst.iso.torrent")
	if err!=nil{
		//then some issue here
		print("Error", err)
	}

	tf.DownloadToFile()

	fmt.Println("eer", err)
}
