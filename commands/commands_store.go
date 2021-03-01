package commands

import (
	"github.com/spf13/cobra"
)

var (
	storeCmd = &cobra.Command{
		Use:   "store <repository>",
		Run:   storeCommand,
		Short: "Store files from the self repository to another repository.",
		Long:  "Create a commit and send new blobs and timeline to another repository.",
		Args:  cobra.ExactArgs(1),
	}
)

func storeCommand(cmd *cobra.Command, args []string) {
	if err := storeAction(args[0]); err != nil {
		Exit(err, 1)
	}
}

func init() {
	RootCmd.AddCommand(storeCmd)
}

func storeAction(repoName string) (err error) {
	// - 現在のディレクトリ構成とファイルのリストを記録 (commit)
	//   - args[0] で指定されたrepository を取得
	remoteRepo, err := findRepo(repoName)
	if err != nil {
		return err
	}

	//   - commitする
	commit, err := createCommit()
	if err != nil {
		return err
	}

	// - store先のhashリストを取得
	remoteHashStrings, err := remoteRepo.FetchBlobHashes()
	if err != nil {
		return err
	}

	// - storeされていないファイルを送信する
	var tagsToSend []Tag
	for _, tag := range commit.Tags {
		if !isIncluded(remoteHashStrings, tag.Hash.String()) {
			tagsToSend = append(tagsToSend, tag)
		}
	}
	err = remoteRepo.sendLocalBlobs(tagsToSend)
	if err != nil {
		return err
	}

	//   - commit を repository の list にマージする
	//     add a commit to repoPath/.arciv/commit
	//     commit write to repoPath/.arciv/list/[commit-id]
	// TODO: Repository.Hogefuga() 側で存在チェックをしてくれたら、こちら側の存在チェックを削除
	remoteTimeline, err := remoteRepo.LoadTimeline()
	if err != nil {
		return err
	}
	if isIncluded(remoteTimeline, commit.Id) {
		message("The commit " + commit.Id + " already exists in the timeline of the repository " + remoteRepo.Name)
		return nil
	}
	err = remoteRepo.WriteTags(commit)
	if err != nil {
		return err
	}
	return remoteRepo.AddTimeline(commit)
}

func isIncluded(strs []string, s string) bool {
	for _, str := range strs {
		if str == s {
			return true
		}
	}
	return false
}
