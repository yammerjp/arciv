package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	//  "github.com/aws/aws-sdk-go-v2"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Op struct {
	findFilePaths func(root string) (relativePaths []string, err error)
	writeLines   func(path string, lines []string) error
	loadLines    func(path string) ([]string, error)
	sendBlobs    func(paths, names []string) error
	receiveBlobs func(paths, names []string) error
	// receiveBlobsRequest func(names []string, validDays int) (restoreNames []string, err error)
}

var s3Op *S3Op

type S3BucketClient struct {
	S3client   *s3.Client
	BucketName string
	RegionName string
}

var s3BucketClient *S3BucketClient

func prepareS3BucketClient(bucketName, regionName string) {
	if s3BucketClient != nil && s3BucketClient.BucketName == bucketName && s3BucketClient.RegionName == regionName {
		return
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(regionName))
	if err != nil {
		panic(err)
	}
	s3BucketClient = &S3BucketClient{
		S3client:   s3.NewFromConfig(cfg),
		BucketName: bucketName,
	}
}

func (bucketClient S3BucketClient) list() (keys []string, err error) {
	p := s3.NewListObjectsV2Paginator(
		bucketClient.S3client,
		&s3.ListObjectsV2Input{
			Bucket: &bucketClient.BucketName,
		},
	)

	var i int
	for p.HasMorePages() {
		i++
		page, err := p.NextPage(context.TODO())
		if err != nil {
			fmt.Println("error occured")
			return []string{}, err
		}
		for _, obj := range page.Contents {
			keys = append(keys, *obj.Key)
		}
	}
	return keys, nil
}

func (bucketClient S3BucketClient) getReader(key string) (io.Reader, error) {
	got, err := bucketClient.S3client.GetObject(
		context.TODO(),
		&s3.GetObjectInput{
			Bucket: &bucketClient.BucketName,
			Key:    &key,
		},
	)
	if err != nil {
		return nil, err
	}
	return got.Body, nil
}

func (bucketClient S3BucketClient) getLines(key string) (lines []string, err error) {
	reader, err := bucketClient.getReader(key)
	if err != nil {
		return []string{}, err
	}
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, nil
}

func (bucketClient S3BucketClient) getFile(key, localPath string) error {
	reader, err := bucketClient.getReader(key)
	if err != nil {
		return err
	}
	file, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	return err
}

func (bucketClient S3BucketClient) putReader(key string, reader io.Reader, storageClass types.StorageClass) error {
	_, err := bucketClient.S3client.PutObject(
		context.TODO(),
		&s3.PutObjectInput{
			Bucket:       &bucketClient.BucketName,
			Key:          &key,
			Body:         reader,
			StorageClass: storageClass,
		},
	)
	return err
}

func (bucketClient S3BucketClient) putLines(key string, lines []string) error {
	return bucketClient.putReader(
		key,
		strings.NewReader(strings.Join(lines, "\n")),
		types.StorageClassStandard,
	)
}

func (bucketClient S3BucketClient) putFile2deepArchive(key, localPath string) error {
	f, err := os.OpenFile(localPath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	return bucketClient.putReader(key, f, types.StorageClassDeepArchive)
}

func (bucketClient S3BucketClient) restoreRequest(key string, restoreKeyRequest string, validDays int32) (restoreKey string, err error) {
	got, err := bucketClient.S3client.RestoreObject(
		context.TODO(),
		&s3.RestoreObjectInput{
			Bucket: &bucketClient.BucketName,
			Key:    &key,
			RestoreRequest: &types.RestoreRequest{
				Days: validDays,
				Tier: types.TierBulk,
				OutputLocation: &types.OutputLocation{
					S3: &types.S3Location{
						BucketName: &restoreKeyRequest,
					},
				},
			},
		},
	)
	if err != nil {
		return "", err
	}
	return *got.RestoreOutputPath, nil
}

func init() {
	s3Op = &S3Op{
		findFilePaths: func(root string) (relativePaths []string, err error) {
			if s3BucketClient == nil {
				return []string{}, errors.New("S3BucketClient is not prepared")
			}
			keys, err := s3BucketClient.list()
			if err != nil {
				return []string{}, err
			}
			for _, key := range keys {
				if strings.HasPrefix(key, root + "/") {
					relativePaths = append(relativePaths, key[len(root)+1:])
				}
			}
			return relativePaths, nil
		},
		writeLines: func(path string, lines []string) error {
			if s3BucketClient == nil {
				return errors.New("S3BucketClient is not prepared")
			}
			return s3BucketClient.putLines(path, lines)
		},
		loadLines: func(path string) ([]string, error) {
			if s3BucketClient == nil {
				return []string{}, errors.New("S3BucketClient is not prepared")
			}
			return s3BucketClient.getLines(path)
		},
		sendBlobs: func(paths, names []string) error {
			if s3BucketClient == nil {
				return errors.New("S3BucketClient is not prepared")
			}
			if len(paths) != len(names) {
				return errors.New("arguments of receiveBlobs() require the same length slice")
			}
			for i, path := range paths {
				err := s3BucketClient.putFile2deepArchive(names[i], path)
				if err != nil {
					return err
				}
			}
			return nil
		},
		receiveBlobs: func(paths, names []string) error {
			if s3BucketClient == nil {
				return errors.New("S3BucketClient is not prepared")
			}
			if len(paths) != len(names) {
				return errors.New("arguments of receiveBlobs() require the same length slice")
			}
			for i, path := range paths {
				err := s3BucketClient.getFile(names[i], path)
				if err != nil {
					return err
				}
			}
			return nil
		},
		/*
			receiveBlobsRequest: func(names []string, validDays int) (restoreNames []string, err error) {
				if s3BucketClient == nil {
					return []string{}, errors.New("S3BucketClient is not prepared")
				}
				for _, name := range names {
					restoreName, err := s3BucketClient.restoreRequest(name)
					if err != nil {
						return restoreNames, err
					}
					restoreNames = append(restoreNames, restoreName)
				}
				return restoreNames, nil
			},
		*/
	}
}
