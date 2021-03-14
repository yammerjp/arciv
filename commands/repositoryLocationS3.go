package commands

type RepositoryLocationS3 struct {
	BucketName string
	RegionName string
}

func (r RepositoryLocationS3) String() string {
	return "type:s3 region:" + r.RegionName + " bucket:" + r.BucketName
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

func (r RepositoryLocationS3) ReceiveRemoteBlobsRequest(tags []Tag) (keysRequested []string, err error) {
	var keys []string
	for _, tag := range tags {
		key := ".arciv/blob/" + tag.Hash.String()
		if !isIncluded(keys, key) {
			keys = append(keys, key)
		}
	}
	return s3Op.receiveBlobsRequest(r.RegionName, r.BucketName, keys, 3)
}

func (r RepositoryLocationS3) ReceiveRemoteBlobs(tags []Tag) (err error) {
	var toPaths []string
	var keys []string
	base := fileOp.rootDir() + "/.arciv/blob/"
	for _, tag := range tags {
		blob := tag.Hash.String()
		toPaths = append(toPaths, base+blob)
		keys = append(keys, ".arciv/blob/"+blob)
	}
	return s3Op.receiveBlobs(r.RegionName, r.BucketName, toPaths, keys)
}
