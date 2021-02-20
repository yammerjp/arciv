package commands

import (
  "github.com/spf13/cobra"
  "fmt"
  "os"
  "strings"
  "errors"
  "path/filepath"
  "crypto/sha256"
  "io"
  "encoding/hex"
)

var (
  commitCmd = &cobra.Command{
     Use: "commit",
     Run: commitCommand,
  }
)

func commitCommand(cmd *cobra.Command, args []string) {
  if err := commitAction(); err != nil {
    Exit(err, 1)
  }
}

func commitAction() (err error) {
  rootDir, err := findRoot()
  if err != nil {
    return err
  }
  paths, err := fullFileList(rootDir)
  if err != nil {
    return err
  }
  for _, path := range(paths) {
//    fmt.Println(path)
    hasher := sha256.New()
    f, err := os.Open(path)
    if err != nil {
      return err
    }
    defer f.Close()
    if _, err := io.Copy(hasher, f); err != nil {
      return err
    }
    fmt.Println(hex.EncodeToString(hasher.Sum(nil)) + " " + path)
  }
  return nil
}

func init() {
  RootCmd.AddCommand(commitCmd)
}

func findRoot() (string, error){
  // find root directory (exist .arciv)
  // ex . current dir is /hoge/fuga/wara
  // search /hoge/fuga/wara/.arciv , and next /hoge/fuga/.arciv , and next /hoge/.arciv , and next /.arciv
  currentDir, _ := os.Getwd()
  for dir := currentDir; strings.LastIndex(dir, "/") != -1; dir = dir[:strings.LastIndex(dir, "/")]{
    if f, err := os.Stat(dir + "/.arciv"); ! os.IsNotExist(err) && f.IsDir() {
      return dir , nil
    }
  }
  return "", errors.New(".arciv is not found")
}

func fullFileList(rootDir string) ([]string, error) {
  var files []string
  err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
    if err != nil {
      return err
    }
    if info.IsDir() && info.Name() == ".arciv" {
      return filepath.SkipDir
    }
    if !info.IsDir() {
      files = append(files, path[len(rootDir)+1:])
    }
    return nil
  })

  if err != nil {
    return []string{}, err
  }
  return files, nil
}
