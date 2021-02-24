package commands

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var (
	repositoriesCmd = &cobra.Command{
		Use: "repositories",
		Run: repositoriesCommand,
	}
)

func repositoriesCommand(cmd *cobra.Command, args []string) {
	if err := repositoriesAction(args); err != nil {
		Exit(err, 1)
	}
}

func init() {
	RootCmd.AddCommand(repositoriesCmd)
}

func repositoriesAction(args []string) (err error) {
	if len(args) == 1 && args[0] == "show" {
		return repositoriesActionShow()
	}
	if len(args) == 3 && args[0] == "add" {
		return repositoriesActionAdd(args[1], args[2])
	}
	if len(args) == 2 && args[0] == "remove" {
		return repositoriesActionRemove(args[1])
	}
	fmt.Fprintln(os.Stderr, "Usage: arciv repositoreis show")
	fmt.Fprintln(os.Stderr, "       arciv repositoreis add [repository name] [repository path]")
	fmt.Fprintln(os.Stderr, "       arciv repositoreis remove [repository name]")
	return nil
}

func repositoriesActionShow() error {
	repos, err := loadRepos()
	if err != nil {
		return err
	}
	for _, repo := range repos {
		fmt.Println(repo)
	}
	return nil
}

func repositoriesActionAdd(name string, url string) error {
	if strings.Index(name, " ") != -1 {
		return errors.New("Repository name must not include space")
	}
	repos, err := loadRepos()
	if err != nil {
		return err
	}
	for _, r := range repos {
		if r.Name == name {
			return errors.New("The repository name already exists")
		}
	}
	repo, err := createRepoStruct(name, url)
	if err != nil {
		return err
	}
	repos = append(repos, repo)
	return reposWrite(repos)
}

func repositoriesActionRemove(name string) error {
	if name == "self" {
		return errors.New("self must be exist")
	}
	repos, err := loadRepos()
	if err != nil {
		return err
	}
	for i, repo := range repos {
		if repo.Name != name {
			continue
		}
		// delete repos[i]
		repos = append(repos[:i], repos[i+1:]...)
		return reposWrite(repos)
	}
	return errors.New("The repository is not found")
}

func loadRepos() ([]Repository, error) {
	lines, err := loadLines(rootDir() + "/.arciv/repositories")
	if err != nil {
		return []Repository{}, err
	}
	repos := []Repository{SelfRepo()}
	for _, line := range lines {
		idx := strings.Index(line, " ")
		if idx == -1 {
			return []Repository{}, errors.New("Repository path is not registerd in .arciv/repositories")
		}
		name := line[:idx]
		url := line[idx+1:]

		for _, r := range repos {
			if r.Name == name {
				return []Repository{}, errors.New("Repositoy name is conflict in .arciv/repositories")
			}
		}
		repo, err := createRepoStruct(name, url)
		if err != nil {
			return []Repository{}, err
		}
		repos = append(repos, repo)
	}
	return repos, nil
}

func createRepoStruct(name string, url string) (Repository, error) {
	var path string
	var pathType PathType
	if strings.HasPrefix(url, "file://") {
		path = url[len("file://"):]
		pathType = PATH_FILE
	} else {
		return Repository{}, errors.New("Repository path must be file:///...")
	}
	return Repository{Name: name, Path: path, PathType: pathType}, nil
}

func findRepo(name string) (Repository, error) {
	repos, err := loadRepos()
	if err != nil {
		return Repository{}, err
	}
	for _, repo := range repos {
		if repo.Name == name {
			return repo, nil
		}
	}
	return Repository{}, errors.New("Repository is not found")
}

func reposWrite(repos []Repository) error {
	root := rootDir()
	err := os.Rename(root+"/.arciv/repositories", root+"/.arciv/repositories.org")
	if err != nil {
		return err
	}
	file, err := os.OpenFile(root+"/.arciv/repositories", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, repo := range repos {
		if repo.Name == "self" {
			// Do not write out self
			continue
		}
		fmt.Fprintln(file, repo.String())
	}

	err = os.Remove(root + "/.arciv/repositories.org")
	if err != nil {
		return err
	}
	return nil
}
