package main

import (
	downloader2 "S3Downloader/downloader"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"os"
)

func main() {
	config, err := downloader2.Configurate()
	var fileHandlingError *downloader2.FileHandlingError
	var flagError *downloader2.FlagError
	if err != nil || config.File == nil {
		if errors.As(err, &fileHandlingError) {
			fmt.Printf("Configuration error by file handling: %s\n", err.Error())
			return
		} else if errors.As(err, &flagError) {
			fmt.Printf("Configuration error by incorrect flags: %s\n", err.Error())
			return
		}
	}
	defer config.File.Close()
	downloader := downloader2.NewS3Downloader(config.Client, 0)
	si3 := &s3.GetObjectInput{
		Bucket: aws.String(config.Bucket),
		Key:    aws.String(config.Name),
	}
	bytes, err := downloader.Download(config.File, si3)
	if err != nil {
		fmt.Printf("Target file downloading isn't successfull: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("%d bytes downloaded\n", bytes)
}
