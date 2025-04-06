package main

import (
	"fmt"
)

func main() {
	config, err := Configurate()
	if err != nil {
	}
	downloader := NewDownloader(config.client, 0)
	bytes, err := downloader.Download(config.file, config.bucket, config.name)
	if err != nil {
	}
	fmt.Printf("%d bytes downloaded\n", bytes)
	config.file.Close()
}
