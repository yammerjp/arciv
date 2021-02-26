package commands

import (
	"github.com/spf13/cobra"
)

var (
	logCmd = &cobra.Command{
		Use: "log",
		Run: logCommand,
    Args: cobra.NoArgs,
	}
)

var repoName string
var commitAlias string

func logCommand(cmd *cobra.Command, args []string) {
	if err := logAction(args); err != nil {
		Exit(err, 1)
	}
}

func init() {
	RootCmd.AddCommand(logCmd)
  logCmd.Flags().StringVarP(&repoName, "repository", "r", "", "repository name")
  logCmd.Flags().StringVarP(&commitAlias, "commit", "c", "", "commit id")
}

func logAction(args []string) (err error) {
  var repo Repository
  if repoName == "" {
    repo = SelfRepo()
  } else {
    repo, err = findRepo(repoName)
    if err != nil {
      return err
    }
  }

  if commitAlias == "" {
    return printTimeline(repo)
  }
  commit, err := repo.LoadCommitFromAlias(commitAlias)
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
	for _, p := range c.Photos {
		messageStdin(p.String())
	}
	return nil
}
