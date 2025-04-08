package downloader

import (
	"errors"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/spf13/pflag"
	"io"
	"os"
	"strings"
	"unicode"
)

//go:generate mockery --name=Downloader --output=./mocks --filename=mock_downloader.go

type FlagError int

type FileHandlingError string

func (e FlagError) Error() string {
	return fmt.Sprintf("Invalid command line flags: [code %d]\n", e)
}

func (e FileHandlingError) Error() string {
	return fmt.Sprintf("Invalid file handling: %s\n", string(e))
}

const (
	INCORRECT_BUCKET_FLAG FlagError = iota + 600
	INCORRECT_NAME_FLAG
	INCORRECT_FILE_FLAG
	INCORRECT_REGION_FLAG
)

const (
	ERROR_IN_FILE_CREATION          FileHandlingError = "error during creating file"
	TARGET_DOWNLOADING_FILE_ABSENCE FileHandlingError = "don't found a necessary file"
)

type ConfigurationStructure struct {
	Client Downloader
	Bucket string
	Name   string
	File   *os.File
}

type S3Downloader struct {
	Client   Downloader
	NumBytes int64
}

type Downloader interface {
	Download(io.WriterAt, *s3.GetObjectInput, ...func(*s3manager.Downloader)) (int64, error)
}

func FlagHandling(ptr *string) error {
	if ptr == nil || *ptr == "" || len(*ptr) > 20 {
		return errors.New("nil pointer")
	}
	if strings.HasPrefix(*ptr, "bucket") && strings.Contains(*ptr, "invalid") {
		return INCORRECT_BUCKET_FLAG
	} else if strings.HasPrefix(*ptr, "name") {
		for _, val := range *ptr {
			if unicode.IsDigit(val) {
				return INCORRECT_NAME_FLAG
			}
		}
	} else if strings.HasPrefix(*ptr, "file") && !strings.HasSuffix(*ptr, ".txt") {
		return INCORRECT_FILE_FLAG
	} else if !strings.HasPrefix(*ptr, "us") {
		return INCORRECT_REGION_FLAG
	}

	return nil
}

func Configurate() (*ConfigurationStructure, error) {
	pflag.CommandLine = pflag.NewFlagSet("", pflag.ContinueOnError)
	bucketFlag := pflag.String("bucket", "bucket_abc", "S3 bucket name")
	nameFlag := pflag.String("name", "name_xyz", "S3 downloader name")
	fileFlag := pflag.String("file", "file_test3.txt", "File to download")
	regionFlag := pflag.String("region", "us-east-1", "S3 bucket region")
	pflag.Parse()

	err1 := FlagHandling(bucketFlag)
	if err1 != nil {
		return &ConfigurationStructure{}, err1
	}

	err2 := FlagHandling(nameFlag)
	if err2 != nil {
		return &ConfigurationStructure{}, err2
	}

	err3 := FlagHandling(fileFlag)
	if err3 != nil {
		return &ConfigurationStructure{}, err3
	}

	err4 := FlagHandling(regionFlag)
	if err4 != nil {
		return &ConfigurationStructure{}, err4
	}

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(*regionFlag),
	},
	)
	downloader := s3manager.NewDownloader(sess)
	f, err := os.Create(*fileFlag)
	if err != nil {
		return &ConfigurationStructure{}, ERROR_IN_FILE_CREATION
	}

	return &ConfigurationStructure{
		Client: downloader,
		Bucket: *bucketFlag,
		Name:   *nameFlag,
		File:   f,
	}, nil
}

func NewS3Downloader(client Downloader, numBytes int64) *S3Downloader {
	return &S3Downloader{
		Client:   client,
		NumBytes: numBytes,
	}
}

func (d *S3Downloader) Download(file *os.File, si3 *s3.GetObjectInput) (int64, error) {
	d.NumBytes = 0
	err := retry.Do(
		func() error {
			var err error
			numBytes, err := d.Client.Download(file, si3)
			if err != nil {
				err = TARGET_DOWNLOADING_FILE_ABSENCE
				fmt.Errorf("Error: %v", err)
				return err
			}
			d.NumBytes = numBytes
			return err
		},
		retry.Attempts(1),
		retry.OnRetry(func(n uint, err error) {
			fmt.Printf("Retrying request after error: %v", err)
		}),
	)
	if err != nil {
		err = TARGET_DOWNLOADING_FILE_ABSENCE
		fmt.Errorf("Error: %v", err)
		return -1, err
	}

	return d.NumBytes, err
}
