package commands

import (
	"github.com/spf13/cobra"
)

var (
	stashCmd = &cobra.Command{
		Use:   "stash",
		Run:   stashCommand,
		Short: "Stash all files in the self repository to .arciv/blob directory",
		Long: `Create a commit and move all files in the self repository to .arciv/blob directory.
You can restore moved files with excuting the subcommand 'unstash', after excuting the subcommand 'stash'`,
		Args: cobra.NoArgs,
	}
)

func stashCommand(cmd *cobra.Command, args []string) {
	if err := stashAction(); err != nil {
		Exit(err, 1)
	}
}

func stashAction() (err error) {
	commit, err := createCommitStructure()
	if err != nil {
		return err
	}

	err = SelfRepo().AddCommit(commit)
	if err != nil {
		return err
	}
	message("created commit '" + commit.Id + "'")

	err = stashTags(commit.Tags)
	if err != nil {
		return err
	}

	message("stashed all files")
	return nil
}

func stashTags(tags []Tag) (err error) {
	root := fileOp.rootDir()

	// move all files to .arciv/blob
	for _, p := range tags {
		from := root + "/" + p.Path
		to := root + "/.arciv/blob/" + p.Hash.String()
		err = fileOp.moveFile(from, to)
		if err != nil {
			return err
		}
		message("moved " + from + " -> " + to)
	}

	// remove all directory in root without .arciv
	dirPaths, err := fileOp.findDirPaths(root)
	if err != nil {
		return err
	}
	for i := len(dirPaths) - 1; i >= 0; i-- {
		fileOp.removeFile(dirPaths[i])
	}
	return nil
}

func init() {
	RootCmd.AddCommand(stashCmd)
	stashCmd.Flags().BoolVarP(&runFastlyOption, "fast", "s", false, "Check fastly with checking timestamp, without checking file hash")
}
