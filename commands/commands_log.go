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
  if len(args) > 2 {
    return errors.New("Usage: arciv log ([repository-name])")
  }
  var repo Repository
  if len(args) == 0 {
    repo = selfRepo
  } else {
    repo, err = findRepo(args[0])
    if err != nil {
      return err
    }
  }

	timeline, err := repo.LoadTimeline()
	if err != nil {
		return err
	}
	for _, cId := range timeline {
		fmt.Println(cId)
	}

	return nil
}

func init() {
	RootCmd.AddCommand(logCmd)
}
