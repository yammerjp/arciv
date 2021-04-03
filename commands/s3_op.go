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
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Op struct {
	findFilePaths       func(region string, bucket string, root string) (relativePaths []string, err error)
	writeLines          func(region string, bucket string, path string, lines []string) error
	loadLines           func(region string, bucket string, path string) ([]string, error)
	sendBlobs           func(region string, bucket string, paths, names []string) error
	receiveBlobs        func(region string, bucket string, paths, names []string) error
	receiveBlobsRequest func(region string, bucket string, names []string, validDays int32) (namesRequested []string, err error)
}

var s3Op *S3Op

type S3BucketClient struct {
	S3client   *s3.Client
	BucketName string
	RegionName string
}

var s3BucketClient *S3BucketClient

func client(region, bucket string) *S3BucketClient {
	if s3BucketClient != nil && s3BucketClient.RegionName == region && s3BucketClient.BucketName == bucket {
		return s3BucketClient
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		panic(err)
	}
	s3BucketClient = &S3BucketClient{
		S3client:   s3.NewFromConfig(cfg),
		RegionName: region,
		BucketName: bucket,
	}
	return s3BucketClient
}

func (bucketClient S3BucketClient) list(prefix *string) (keys []string, err error) {
	p := s3.NewListObjectsV2Paginator(
		bucketClient.S3client,
		&s3.ListObjectsV2Input{
			Bucket: &bucketClient.BucketName,
			Prefix: prefix,
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

func (bucketClient S3BucketClient) getLines(key string) (lines []string, err error) {
	got, err := bucketClient.S3client.GetObject(
		context.TODO(),
		&s3.GetObjectInput{
			Bucket: &bucketClient.BucketName,
			Key:    &key,
		},
	)
	if err != nil {
		return []string{}, err
	}
	defer got.Body.Close()
	scanner := bufio.NewScanner(got.Body)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, nil
}

func (bucketClient S3BucketClient) getFile(key, localPath string) error {
	got, err := bucketClient.S3client.GetObject(
		context.TODO(),
		&s3.GetObjectInput{
			Bucket: &bucketClient.BucketName,
			Key:    &key,
		},
	)
	if err != nil {
		return err
	}
	defer got.Body.Close()
	file, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, got.Body)
	return err
}

func (bucketClient S3BucketClient) putLines(key string, lines []string) error {
	_, err := bucketClient.S3client.PutObject(
		context.TODO(),
		&s3.PutObjectInput{
			Bucket:       &bucketClient.BucketName,
			Key:          &key,
			Body:         strings.NewReader(strings.Join(lines, "\n")),
			StorageClass: types.StorageClassStandard,
		},
	)
	return err
}

func (bucketClient S3BucketClient) putFile2deepArchive(key, localPath string) error {
	f, err := os.OpenFile(localPath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	uploader := manager.NewUploader(bucketClient.S3client, func(u *manager.Uploader) {
		u.PartSize = 100 * 1024 * 1024 // 100MB par part
		u.Concurrency = 10
	})
	_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:       &bucketClient.BucketName,
		Key:          &key,
		Body:         f,
		StorageClass: types.StorageClassDeepArchive,
	},
	)
	return err
}

func (bucketClient S3BucketClient) restoreRequest(key string, validDays int32) error {
	_, err := bucketClient.S3client.RestoreObject(
		context.TODO(),
		&s3.RestoreObjectInput{
			Bucket: &bucketClient.BucketName,
			Key:    &key,
			RestoreRequest: &types.RestoreRequest{
				Days: validDays,
				GlacierJobParameters: &types.GlacierJobParameters{
					Tier: "Bulk",
				},
			},
		},
	)
	return err
}

func init() {
	s3Op = &S3Op{
		findFilePaths: func(region string, bucket string, root string) (relativePaths []string, err error) {
			if len(root) == 0 {
				return client(region, bucket).list(nil)
			}
			prefix := root + "/"
			keys, err := client(region, bucket).list(&prefix)
			if err != nil {
				return []string{}, err
			}
			for _, key := range keys {
				relativePaths = append(relativePaths, key[len(root)+1:])
			}
			return relativePaths, nil
		},
		writeLines: func(region string, bucket string, path string, lines []string) error {
			return client(region, bucket).putLines(path, lines)
		},
		loadLines: func(region string, bucket string, path string) ([]string, error) {
			return client(region, bucket).getLines(path)
		},
		sendBlobs: func(region string, bucket string, paths, names []string) error {
			if len(paths) != len(names) {
				return errors.New("arguments of receiveBlobs() require the same length slice")
			}
			for i, path := range paths {
				err := client(region, bucket).putFile2deepArchive(names[i], path)
				if err != nil {
					return err
				}
				message("Uploaded: " + path + " (file) -> " + names[i] + " (s3)")
			}
			return nil
		},
		receiveBlobs: func(region string, bucket string, paths, names []string) error {
			if len(paths) != len(names) {
				return errors.New("arguments of receiveBlobs() require the same length slice")
			}
			for i, path := range paths {
				err := client(region, bucket).getFile(names[i], path+".download")
				if err != nil {
					return err
				}
				err = fileOp.moveFile(path+".download", path)
				if err != nil {
					return err
				}
				message("Downloaded: " + names[i] + " (s3) -> " + path + " (file)")
			}
			return nil
		},
		receiveBlobsRequest: func(region string, bucket string, names []string, validDays int32) (namesRequested []string, err error) {
			for i, name := range names {
				err := client(region, bucket).restoreRequest(name, validDays)
				if err != nil {
					return names[:i], err
				}
			}
			return names, nil
		},
	}
}
