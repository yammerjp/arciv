package commands

type RepositoryLocationFile struct {
	Path string
}

func (repositoryLocationFile RepositoryLocationFile) String() string {
	return "file://" + repositoryLocationFile.Path
}

func (repositoryLocationFile RepositoryLocationFile) writeLines(relativePath string, lines []string) error {
	return fileOp.writeLines(repositoryLocationFile.Path+"/"+relativePath, lines)
}

func (repositoryLocationFile RepositoryLocationFile) loadLines(relativePath string) (lines []string, err error) {
	return fileOp.loadLines(repositoryLocationFile.Path + "/" + relativePath)
}

func (repositoryLocationFile RepositoryLocationFile) findFilePaths(root string) (relativePaths []string, err error) {
	return fileOp.findFilePaths(repositoryLocationFile.Path + "/" + root)
}

func (repositoryLocationFile RepositoryLocationFile) SendLocalBlobs(tags []Tag) (err error) {
	for _, tag := range tags {
		from := fileOp.rootDir() + "/" + tag.Path
		to := repositoryLocationFile.Path + "/.arciv/blob/" + tag.Hash.String()
		err = fileOp.copyFile(from, to)
		if err != nil {
			return err
		}
		message("uploaded: " + tag.Hash.String() + ", " + tag.Path)
	}
	return nil
}

func (repositoryLocationFile RepositoryLocationFile) ReceiveRemoteBlobs(tags []Tag) (err error) {
	for _, tag := range tags {
		from := repositoryLocationFile.Path + "/.arciv/blob/" + tag.Hash.String()
		to := fileOp.rootDir() + "/.arciv/blob/" + tag.Hash.String()
		err = fileOp.copyFile(from, to)
		if err != nil {
			return err
		}
		message("downloaded: " + tag.Hash.String() + ", will locate to: " + tag.Path)
	}
	return nil
}
