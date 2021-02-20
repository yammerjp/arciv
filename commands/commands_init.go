package commands

import (
  "github.com/spf13/cobra"
  "fmt"
)

var (
  initCmd = &cobra.Command{
     Use: "init",
     Run: initCommand,
  }
)

func initCommand(cmd *cobra.Command, args []string) {
  if err := initAction(); err != nil {
    Exit(err, 1)
  }
}

func initAction() (err error) {
  fmt.Println("hello, arciv")
  return nil
}

func init() {
  RootCmd.AddCommand(initCmd)
}
