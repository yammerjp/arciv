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

func (r Repository) Init() error {
	createDirsInDotArciv := []string{"list", "blob"}
	switch lf := r.Location.(type) {
	case RepositoryLocationFile:
		for _, dir := range createDirsInDotArciv {
			err := fileOp.mkdirAll(lf.Path + "/.arciv/" + dir)
			if err != nil {
				return err
			}
		}
	}

	createFilesInDotArciv := []string{"repositories", "timeline", "timestamps"}
	paths, err := r.Location.findFilePaths(".arciv")
	if err != nil {
		return err
	}
	for _, file := range createFilesInDotArciv {
		if isIncluded(paths, file) {
			continue
		}
		err := r.Location.writeLines(".arciv/"+file, []string{})
		if err != nil {
			return err
		}
	}
	return nil
}
func init() {
	RootCmd.AddCommand(initCmd)
}
