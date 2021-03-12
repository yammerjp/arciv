package commands

import (
	"errors"
)

type RepositoryLocationS3 struct {
	BucketName string
	RegionName string
}

func (repositoryLocationS3 RepositoryLocationS3) String() string {
	return "s3://" + repositoryLocationS3.BucketName
}

func (repositoryLocationS3 RepositoryLocationS3) writeLines(relativePath string, lines []string) error {
	repositoryLocationS3.prepareClient()
	return s3Op.writeLines(relativePath, lines)
}

func (repositoryLocationS3 RepositoryLocationS3) loadLines(relativePath string) (lines []string, err error) {
	repositoryLocationS3.prepareClient()
	return s3Op.loadLines(relativePath)
}

func (repositoryLocationS3 RepositoryLocationS3) findFilePaths(root string) (relativePaths []string, err error) {
	if root != ".arciv/blob" {
		panic("findFilePaths() out of '.arciv/blob/ is not implemented")
	}
	repositoryLocationS3.prepareClient()
	return s3Op.listBlobs()
}

func (repositoryLocationS3 RepositoryLocationS3) SendLocalBlobs(tags []Tag) (err error) {
	var fromPaths []string
	var blobNames []string
	for _, tag := range tags {
		fromPaths = append(fromPaths, fileOp.rootDir()+"/"+tag.Path)
		blobNames = append(blobNames, tag.Hash.String())
	}
	repositoryLocationS3.prepareClient()
	return s3Op.sendBlobs(fromPaths, blobNames)
}

func (repositoryLocationS3 RepositoryLocationS3) ReceiveRemoteBlobs(tags []Tag) (err error) {
	// FIXME: receive restored files from deep archive
	return errors.New("Download blobs from AWS S3 is not implemented yet...\n Please download from Web console and place the files into .arciv/blob/")
	/*
		var toPaths []string
		var blobNames []string
		for _, tag := range tags {
			toPaths = append(toPaths, repository.Path+"/.arciv/blob/"+tag.Hash.String())
			blobNames = append(blobNames, tag.Hash.String())
		}
		repositoryLocationS3.prepareClient()
		return s3Op.receiveBlobs(toPaths, blobNames)
	*/
}

func (repositoryLocationS3 RepositoryLocationS3) prepareClient() {
	prepareS3BucketClient(repositoryLocationS3.BucketName, repositoryLocationS3.RegionName)
}