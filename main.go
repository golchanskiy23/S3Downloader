package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	config, err := Configurate()
	if err != nil {
		fmt.Errorf("Configuration error: %s", err.Error())
	}
	downloader := NewS3Downloader(config.client, 0)
	si3 := &s3.GetObjectInput{
		Bucket: aws.String(config.bucket),
		Key:    aws.String(config.name),
	}
	bytes, err := downloader.Download(config.file, si3)
	if err != nil {
		fmt.Errorf("Target file downloading isn't successfull: %s", err.Error())
	}
	fmt.Printf("%d bytes downloaded\n", bytes)
	config.file.Close()
}
