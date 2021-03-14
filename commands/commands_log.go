package commands

import (
	"github.com/spf13/cobra"
)

var (
	logCmd = &cobra.Command{
		Use:   "log",
		Run:   logCommand,
		Short: "Print a timeline or a commit",
		Long:  "Print A timeline (of the self repository by default) if --commit option is not set. Print the commit if --commit option is set.",
		Args:  cobra.NoArgs,
	}
)

var repositoryNameOption string
var commitAliasOption string

func logCommand(cmd *cobra.Command, args []string) {
	if err := logAction(args); err != nil {
		Exit(err, 1)
	}
}

func init() {
	RootCmd.AddCommand(logCmd)
	logCmd.Flags().StringVarP(&repositoryNameOption, "repository", "r", "", "repository name")
	logCmd.Flags().StringVarP(&commitAliasOption, "commit", "c", "", "commit id")
}

func logAction(args []string) (err error) {
	var repo Repository
	if repositoryNameOption == "" {
		repo = SelfRepo()
	} else {
		repo, err = findRepo(repositoryNameOption)
		if err != nil {
			return err
		}
	}

	if commitAliasOption == "" {
		return printTimeline(repo)
	}
	commit, err := repo.LoadCommitFromAlias(commitAliasOption)
	if err != nil {
		return err
	}
	return printCommit(commit)
}

func printTimeline(repo Repository) error {
	timeline, err := repo.LoadTimeline()
	if err != nil {
		return err
	}
	for _, cId := range timeline {
		messageStdin(cId)
	}
	return nil
}

func printCommit(c Commit) error {
	for _, p := range c.Tags {
		messageStdin(p.String())
	}
	return nil
}
