package main

import (
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"os"
)

type ConfigurationStructure struct {
	client *s3manager.Downloader
	bucket string
	name   string
	file   *os.File
}

type Downloader struct {
	client   *s3manager.Downloader
	numBytes uint
}

func Configurate() (ConfigurationStructure, error) {
	return ConfigurationStructure{
		client: nil,
		bucket: "",
		name:   "",
		file:   nil,
	}, nil
}

func NewDownloader(client *s3manager.Downloader, numBytes uint) *Downloader {
	return &Downloader{
		client:   client,
		numBytes: numBytes,
	}
}

func (d *Downloader) Download(file *os.File, bucket, name string) (uint, error) {
	return 0, nil
}
