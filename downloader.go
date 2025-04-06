package main

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/spf13/pflag"
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
	bucketFlag := pflag.String("bucket", "", "S3 bucket name")
	nameFlag := pflag.String("name", "", "S3 downloader name")
	fileFlag := pflag.String("file", "", "File to download")
	pflag.Parse()

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	},
	)
	downloader := s3manager.NewDownloader(sess)
	f, err := os.Create(*fileFlag)
	if err != nil {
		return ConfigurationStructure{}, fmt.Errorf("failed to create file %s, %v", *fileFlag, err)
	}

	return ConfigurationStructure{
		client: downloader,
		bucket: *bucketFlag,
		name:   *nameFlag,
		file:   f,
	}, nil
}

func NewDownloader(client *s3manager.Downloader, numBytes uint) *Downloader {
	return &Downloader{
		client:   client,
		numBytes: numBytes,
	}
}

func (d *Downloader) Download(file *os.File, bucket, name string) (uint, error) {
	d.numBytes = 0
	si3 := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(name),
	}

	err := retry.Do(func() error {
		n, err1 := d.client.Download(file, si3)
		if err1 != nil {
			return err1
		}
		d.numBytes = uint(n)
		return err1
	})

	return d.numBytes, err
}
