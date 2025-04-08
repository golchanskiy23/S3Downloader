package downloader

import (
	"S3Downloader/mocks"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/pflag"
	_ "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestFlagHandling(t *testing.T) {
	tests := []struct {
		flagValue   string
		expectedErr error
	}{
		{"bucket_invalid", INCORRECT_BUCKET_FLAG},
		{"name123", INCORRECT_NAME_FLAG},
		{"file_test.pdf", INCORRECT_FILE_FLAG},
		{"region", INCORRECT_REGION_FLAG},
		{"us-east-1", nil},
	}

	for _, tt := range tests {
		t.Run(tt.flagValue, func(t *testing.T) {
			err := FlagHandling(&tt.flagValue)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestConfigurate(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectErr   error
		expectValue string
	}{
		{
			name:        "Valid flags",
			args:        []string{"--bucket=bucket_abc", "--name=name_xyz", "--file=testfile.txt", "--region=us-east-1"},
			expectErr:   nil,
			expectValue: "bucket_abc",
		},
		{
			name:        "File creation error",
			args:        []string{"--bucket=bucket_abc", "--name=name_xyz", "--file=/invalid/path/testfile.txt", "--region=us-east-1"},
			expectErr:   ERROR_IN_FILE_CREATION,
			expectValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pflag.CommandLine = pflag.NewFlagSet("", pflag.ContinueOnError)
			os.Args = append([]string{"cmd"}, tt.args...)

			conf, err := Configurate()

			if tt.expectErr != nil {
				assert.Equal(t, tt.expectErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectValue, conf.Bucket)
			}
		})
	}
}

func generateRandomBytes(size int) []byte {
	rand.Seed(time.Now().UnixNano())
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		panic("failed to generate random bytes")
	}
	return bytes
}

func TestS3DownloaderDownload(t *testing.T) {
	mockDownloader := mocks.NewDownloader(t)
	mockDownloader.On("Download", mock.Anything, mock.Anything, mock.Anything).Return(int64(100), nil)

	downloader := NewS3Downloader(mockDownloader, 0)
	s3Input := &s3.GetObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("test-key"),
	}

	t.Run("successful download", func(t *testing.T) {
		file, _ := os.Create("test-file")
		file.Write(generateRandomBytes(100))
		bytesDownloaded, err := downloader.Download(file, s3Input)
		assert.Nil(t, err)
		assert.Equal(t, int64(100), bytesDownloaded)
		mockDownloader.AssertExpectations(t)
	})

	t.Run("download failed", func(t *testing.T) {
		s := mocks.NewDownloader(t)
		down := NewS3Downloader(s, 0)
		file, err := os.Create("test-file")
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		defer file.Close()

		s.On("Download", file, s3Input).Return(int64(0), TARGET_DOWNLOADING_FILE_ABSENCE).Times(1)

		_, err = down.Download(file, s3Input)
		assert.Error(t, err)
		assert.Equal(t, TARGET_DOWNLOADING_FILE_ABSENCE, err)
		assert.Equal(t, int64(0), down.NumBytes)
		s.AssertExpectations(t)
	})

}
