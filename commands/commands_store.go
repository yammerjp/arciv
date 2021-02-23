package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var (
	storeCmd = &cobra.Command{
		Use: "store",
		Run: storeCommand,
	}
)

func storeCommand(cmd *cobra.Command, args []string) {
	if err := storeAction(args); err != nil {
		Exit(err, 1)
	}
}

func init() {
	RootCmd.AddCommand(storeCmd)
}

func storeAction(args []string) (err error) {
	// - 現在のディレクトリ構成とファイルのリストを記録 (commit)
	//   - args[0] で指定されたrepository を取得
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Usage: arciv store [repository-name]")
		return nil
	}
	remoteRepo, err := findRepo(args[0])
	if err != nil {
		return err
	}

	//   - commitする
	commit, err := createCommit()
	if err != nil {
		return err
	}

	// - store先のhashリストを取得
	remoteHashStrings, err := remoteRepo.fetchBlobHashes()
	if err != nil {
		return err
	}

	// - storeされていないファイルを送信する
	var photosToSend []Photo
	for _, photo := range commit.Photos {
		if !isInclude(remoteHashStrings, photo.Hash.String()) {
			photosToSend = append(photosToSend, photo)
		}
	}
	err = remoteRepo.sendLocalBlobs(photosToSend)
	if err != nil {
		return err
	}

	//   - commit を repository の list にマージする
	//     add a commit to repoPath/.arciv/commit
	//     commit write to repoPath/.arciv/list/[commit-id]
	// TODO: Repository.Hogefuga() 側で存在チェックをしてくれたら、こちら側の存在チェックを削除
	remoteTimeline, err := remoteRepo.loadTimeline()
	if err != nil {
		return err
	}
	if isInclude(remoteTimeline, commit.Id) {
		fmt.Fprintln(os.Stderr, "The commit "+commit.Id+" already exists in the timeline of the repository "+remoteRepo.Name)
		return nil
	}
	err = remoteRepo.WritePhotos(commit)
	if err != nil {
		return err
	}
	return remoteRepo.AddTimeline(commit)
}

func isInclude(strs []string, s string) bool {
	for _, str := range strs {
		if str == s {
			return true
		}
	}
	return false
}
