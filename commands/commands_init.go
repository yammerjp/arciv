package commands

import (
	"errors"
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
	return Repository{Name: "self", Path: currentDir, PathType: PATH_FILE}.Init()
}

func (repository Repository) Init() error {
	if repository.PathType != PATH_FILE {
		return errors.New("Repository's PathType must be PATH_FILE")
	}

	err := os.MkdirAll(repository.Path+"/.arciv/list", 0777)
	if err != nil {
		return err
	}
	err = os.MkdirAll(repository.Path+"/.arciv/blob", 0777)
	if err != nil {
		return err
	}

	fR, err := os.Create(repository.Path + "/.arciv/repositories")
	if err != nil {
		return err
	}
	defer fR.Close()

	fC, err := os.Create(repository.Path + "/.arciv/timeline")
	if err != nil {
		return err
	}
	defer fC.Close()
	return nil
}

func init() {
	RootCmd.AddCommand(initCmd)
}
