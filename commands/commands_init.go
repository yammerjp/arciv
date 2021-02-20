package commands

import (
	"fmt"
	"github.com/spf13/cobra"
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
	err = mkdir(".arciv")
	if err != nil {
		return err
	}
	err = mkdir(".arciv/commit")
	if err != nil {
		return err
	}
	err = mkdir(".arciv/list")
	if err != nil {
		return err
	}
	return nil
}

func mkdir(path string) error {
	err := os.Mkdir(path, 0777)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create directory '"+path+"'")
		return err
	}
	return nil
}

func init() {
	RootCmd.AddCommand(initCmd)
}
