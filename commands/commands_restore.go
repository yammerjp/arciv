package commands

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var (
	restoreCmd = &cobra.Command{
		Use: "restore",
		Run: restoreCommand,
	}
)

func restoreCommand(cmd *cobra.Command, args []string) {
	if err := restoreAction(args); err != nil {
		Exit(err, 1)
	}
}

func restoreAction(args []string) (err error) {
	selfRepo := SelfRepo()
	var repoName, commitAlias string
	dryRun := false
	switch len(args) {
	case 2:
		repoName = args[0]
		commitAlias = args[1]
	case 3:
		dryRun = true
		if args[0] == "dry-run" {
			repoName = args[1]
			commitAlias = args[2]
		} else if args[1] == "dry-run" {
			repoName = args[0]
			commitAlias = args[2]
		} else if args[2] == "dry-run" {
			repoName = args[0]
			commitAlias = args[1]
		} else {
			return errors.New("Usage: arciv restore [repository-name] [alias]")
		}
	default:
		return errors.New("Usage: arciv restore [repository-name] [alias]")
	}
	// fetch remoteCommit
	remoteRepo, err := findRepo(repoName)
	if err != nil {
		return err
	}
	remoteCommit, err := remoteRepo.LoadCommitFromAlias(commitAlias)
	if err != nil {
		return err
	}

	// check no changes
	localCommit, err := createCommitStructure()
	if err != nil {
		return err
	}
	localLatestCommitId, err := selfRepo.LoadLatestCommitId()
	if err != nil {
		return err
	}
	if localCommit.Id[9:] != localLatestCommitId[9:] {
		return errors.New("Directory structure is not saved with latest commit")
	}

	localHashStrings, err := selfRepo.FetchBlobHashes()
	if err != nil {
		return err
	}
	for _, lPhoto := range localCommit.Photos {
		localHashStrings = append(localHashStrings, lPhoto.Hash.String())
	}
	// filter blob hashes to recieve
	var blobsToRecieve []Photo
	for _, rPhoto := range remoteCommit.Photos {
		if !isIncluded(localHashStrings, rPhoto.Hash.String()) {
			blobsToRecieve = append(blobsToRecieve, rPhoto)
		}
	}

	// download
	if dryRun {
		fmt.Fprintln(os.Stderr, "Dry run")
		for _, b := range blobsToRecieve {
			fmt.Fprintln(os.Stderr, "Download: "+b.Hash.String())
		}
	} else {
		err = remoteRepo.ReceiveRemoteBlobs(blobsToRecieve)
		if err != nil {
			return err
		}
	}

	// mv all local files to .arciv/blob
	os.MkdirAll(selfRepo.Path+"/.arciv/blob", 0777)
	for _, lPhoto := range localCommit.Photos {
		from := selfRepo.Path + "/" + lPhoto.Path
		to := selfRepo.Path + "/.arciv/blob/" + lPhoto.Hash.String()
		if dryRun {
			fmt.Fprintf(os.Stderr, "move %s -> %s\n", from, to)
			continue
		}
		err = os.Rename(from, to)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "moved %s -> %s\n", from, to)
	}

	// remove garbages
	if !dryRun {
		paths, err := findPaths(selfRepo.Path, []string{".arciv"}, true)
		if err != nil {
			return err
		}
		for i := len(paths) - 1; i >= 0; i-- {
			err = os.Remove(paths[i])
			if err != nil {
				return err
			}
		}
	}

	// rename
	for _, rPhoto := range remoteCommit.Photos {
		from := selfRepo.Path + "/.arciv/blob/" + rPhoto.Hash.String()
		to := selfRepo.Path + "/" + rPhoto.Path
		if dryRun {
			fmt.Fprintf(os.Stderr, "move %s -> %s\n", from, to)
			continue
		}
		err = os.Rename(from, to)
		if err != nil {
			// mkdirAll and retry
			err = os.MkdirAll(filepath.Dir(to), 0777)
			if err != nil {
				return err
			}
			err = os.Rename(from, to)
			if err != nil {
				return err
			}
		}
		fmt.Fprintf(os.Stderr, "moved %s -> %s\n", from, to)
	}

	if dryRun {
		return nil
	}

	// remove local .arciv/blob/*
	localHashStrings, err = selfRepo.FetchBlobHashes()
	if err != nil {
		return err
	}
	for _, blob := range localHashStrings {
		err = os.Remove(selfRepo.Path + "/.arciv/blob/" + blob)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "removed %s", selfRepo.Path+"/.arciv/blob/"+blob)
	}
	return nil
	//   - restoreするcommitに必要なhashリストを確認
	// - 現在のディレクトリ構成とファイルが既にcommitされたどれかに一致することを確認 (誤った上書きを避ける) (status) --force でスキップ
	// - store先のhashリストを取得
	//   - storeと一緒
	// - .arciv/blob/ 以下にあるファイルをリスト化する
	//   - storeと一緒
	// - 手元のファイルをリネームして、.arciv/blob/以下含め手元に足りないものをダウンロード
	//   - restore する commit にあるhashリストのうち、手元のファイルと手元の .arciv/blob に無いファイルを、手元の.arciv/blobにダウンロード
	//   - 手元のファイルをリネーム
	//   - .arciv/blob のファイルをリネーム
}

func init() {
	RootCmd.AddCommand(restoreCmd)
}
