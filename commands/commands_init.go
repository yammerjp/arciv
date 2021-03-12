package commands

import (
	"github.com/spf13/cobra"
	"os"
)

var (
	initCmd = &cobra.Command{
		Use:   "init",
		Run:   initCommand,
		Short: "Initialize a repository",
		Long: `Initialize a repository.
The repository's root directory specifies the current directory by generating '.arciv' directory on the current directory.`,
		Args: cobra.NoArgs,
	}
)

func initCommand(cmd *cobra.Command, args []string) {
	if err := initAction(); err != nil {
		Exit(err, 1)
	}
}

func initAction() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}
	return Repository{Name: "self", Location: RepositoryLocationFile{Path: currentDir}}.Init()
}

func (repository Repository) Init() error {
	return repository.Location.Init()
}
func (repositoryLocationFile RepositoryLocationFile) Init() error {
	// Create dirs if there does not exist
	err := fileOp.mkdirAll(repositoryLocationFile.Path + "/.arciv/list")
	if err != nil {
		return err
	}
	err = fileOp.mkdirAll(repositoryLocationFile.Path + "/.arciv/blob")
	if err != nil {
		return err
	}

	// Create files if there does not exist
	paths, err := fileOp.findFilePaths(repositoryLocationFile.Path + "/.arciv")
	if err != nil {
		return err
	}
	if !isIncluded(paths, "repositories") {
		fileOp.writeLines(repositoryLocationFile.Path+"/.arciv/repositories", []string{})
	}
	if !isIncluded(paths, "timeline") {
		fileOp.writeLines(repositoryLocationFile.Path+"/.arciv/timeline", []string{})
	}
	if !isIncluded(paths, "timestamps") {
		fileOp.writeLines(repositoryLocationFile.Path+"/.arciv/timestamps", []string{})
	}
	return nil
}

func (r RepositoryLocationS3) Init() error {
	// FIXME: add a func, s3Op.Exist("  key string  ")
	_, err := s3Op.loadLines(r.RegionName, r.BucketName, ".arciv/repositories")
	if err != nil {
		err = s3Op.writeLines(r.RegionName, r.BucketName, ".arciv/repositories", []string{})
		if err != nil {
			return err
		}
	}
	_, err = s3Op.loadLines(r.RegionName, r.BucketName, ".arciv/timeline")
	if err != nil {
		err = s3Op.writeLines(r.RegionName, r.BucketName, ".arciv/timeline", []string{})
		if err != nil {
			return err
		}
	}
	_, err = s3Op.loadLines(r.RegionName, r.BucketName, ".arciv/timestamps")
	if err != nil {
		err = s3Op.writeLines(r.RegionName, r.BucketName, ".arciv/timestamps", []string{})
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	RootCmd.AddCommand(initCmd)
}
