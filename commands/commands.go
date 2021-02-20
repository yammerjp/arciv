package commands

import (
  "fmt"
  "os"
  "github.com/spf13/cobra"
)

var (
  RootCmd = &cobra.Command{
    Use: "arciv",
    Run: func(cmd *cobra.Command, args []string) {
      cmd.Usage()
    },
  }
)

func Run() {
  RootCmd.Execute()
}

func Exit(err error, codes ...int) {
  var code int
  if len(codes) > 0 {
    code = codes[0]
  } else {
    code = 2
  }
  if err != nil {
    fmt.Println(err)
  }
  os.Exit(code)
}
