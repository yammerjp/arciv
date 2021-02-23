package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
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
	remoteRoot, err := remoteRepo.LocalPath()
	if err != nil {
		return err
	}

	//   - commitする
	commit, err := createCommit()
	if err != nil {
		return err
	}

	// - store先のhashリストを取得
	remoteHashStrings, err := fetchRepoHashs(remoteRepo)
	if err != nil {
		return err
	}

	// - storeされていないファイルを送信する
	//   - commit と repository の .arciv/blob のファイル一覧を比較して転送するファイルを決める
	var photosToSend []Photo
	for _, photo := range commit.Photos {
		if isInclude(remoteHashStrings, photo.Hash.String()) {
			continue
		}
		photosToSend = append(photosToSend, photo)
	}

	//   - ファイルを転送する
	err = sendBlobs(remoteRepo, photosToSend)
	if err != nil {
		return err
	}

	//   - commit を repository の list にマージする
	//     add a commit to repoPath/.arciv/commit
	//     commit write to repoPath/.arciv/list/[commit-id]
	remoteTimeline, err := remoteRepo.loadTimeline()
	if err != nil {
		return err
	}
	if isInclude(remoteTimeline, commit.Id) {
		fmt.Fprintln(os.Stderr, "The commit "+commit.Id+" already exists in the timeline of the repository "+remoteRepo.Name)
		return nil
	}
	err = remoteRepo.AddTimeline(commit)
	if err != nil {
		return err
	}
	os.MkdirAll(remoteRoot+"/.arciv/list", 0777)
	err = remoteRepo.WritePhotos(commit)
	if err != nil {
		return err
	}
	return nil
}

func fetchRepoHashs(repo Repository) ([]string, error) {
	//   - .arciv/blob が無ければ掘る
	repoPath, err := repo.LocalPath()
	if err != nil {
		return []string{}, err
	}
	os.MkdirAll(repoPath+"/.arciv/blob", 0777)

	//   - repository の .arciv/blob のファイル一覧を取得する
	repoHashStrings, err := findPaths(repoPath+"/.arciv/blob", []string{})
	if err != nil {
		return []string{}, err
	}
	return repoHashStrings, nil
}

func isInclude(strs []string, s string) bool {
	for _, str := range strs {
		if str == s {
			return true
		}
	}
	return false
}

func sendBlobs(toRepo Repository, photos []Photo) error {
	root := rootDir()
	for _, photo := range photos {
		from := root + "/" + photo.Path
		remoteLocalPath, err := toRepo.LocalPath()
		if err != nil {
			return err
		}
		to := remoteLocalPath + "/.arciv/blob/" + photo.Hash.String()
		err = copyFile(from, to)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "copied %s -> %s\n", from, to)
	}
	return nil
}

func copyFile(from string, to string) error {
	w, err := os.Create(to)
	if err != nil {
		return err
	}
	defer w.Close()

	r, err := os.Open(from)
	if err != nil {
		return err
	}
	defer r.Close()

	_, err = io.Copy(w, r)
	return err
}
