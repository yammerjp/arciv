package commands

import (
	"errors"
)

type RepositoryLocationS3 struct {
	BucketName string
	RegionName string
}

func (r RepositoryLocationS3) String() string {
	return "s3://" + r.BucketName
}

func (r RepositoryLocationS3) writeLines(relativePath string, lines []string) error {
	return s3Op.writeLines(r.RegionName, r.BucketName, relativePath, lines)
}

func (r RepositoryLocationS3) loadLines(relativePath string) (lines []string, err error) {
	return s3Op.loadLines(r.RegionName, r.BucketName, relativePath)
}

func (r RepositoryLocationS3) findFilePaths(root string) (relativePaths []string, err error) {
	return s3Op.findFilePaths(r.RegionName, r.BucketName, root)
}

func (r RepositoryLocationS3) SendLocalBlobs(tags []Tag) (err error) {
	var fromPaths []string
	var blobNames []string
	for _, tag := range tags {
		fromPaths = append(fromPaths, fileOp.rootDir()+"/"+tag.Path)
		blobNames = append(blobNames, tag.Hash.String())
	}
	return s3Op.sendBlobs(r.RegionName, r.BucketName, fromPaths, blobNames)
}

func (r RepositoryLocationS3) ReceiveRemoteBlobs(tags []Tag) (err error) {
	// FIXME: receive restored files from deep archive
	return errors.New("Download blobs from AWS S3 is not implemented yet...\n Please download from Web console and place the files into .arciv/blob/")
	/*
		var toPaths []string
		var blobNames []string
		for _, tag := range tags {
			toPaths = append(toPaths, repository.Path+"/.arciv/blob/"+tag.Hash.String())
			blobNames = append(blobNames, tag.Hash.String())
		}
		return s3Op.receiveBlobs(r.RegionName, r.BucketName, toPaths, blobNames)
	*/
}
