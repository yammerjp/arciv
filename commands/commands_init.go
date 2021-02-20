package commands

import (
  "github.com/spf13/cobra"
  "fmt"
  "os"
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
  if err := os.Mkdir(".arciv", 0777); err != nil {
    fmt.Fprintln(os.Stderr, "Failed to create directory '.arciv'")
    Exit(err, 1)
  }
  return nil
}

func init() {
  RootCmd.AddCommand(initCmd)
}
