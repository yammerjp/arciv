package commands

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	logCmd = &cobra.Command{
		Use: "log",
		Run: logCommand,
	}
)

func logCommand(cmd *cobra.Command, args []string) {
	if err := logAction(args); err != nil {
		Exit(err, 1)
	}
}

func logAction(args []string) (err error) {
	switch len(args) {
	case 0:
		return printTimeline(SelfRepo())
	case 1:
		repo, err := findRepo(args[0])
		if err == nil {
			return printTimeline(repo)
		}
		// FIXME: check error type
		commit, err := SelfRepo().LoadCommitFromAlias(args[0])
		if err != nil {
			return err
		}
		return printCommit(commit)
	case 2:
		repo, err := findRepo(args[0])
		if err != nil {
			return err
		}
		commit, err := repo.LoadCommitFromAlias(args[1])
		if err != nil {
			return err
		}
		return printCommit(commit)
	default:
		return errors.New("Usage: arciv log ([repository-name]) ([commit-alias])")
	}
}

func init() {
	RootCmd.AddCommand(logCmd)
}

func printTimeline(repo Repository) error {
	timeline, err := repo.LoadTimeline()
	if err != nil {
		return err
	}
	for _, cId := range timeline {
		fmt.Println(cId)
	}
	return nil
}

func printCommit(c Commit) error {
	for _, p := range c.Photos {
		fmt.Println(p.String())
	}
	return nil
}
