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
		return reposShow()
	}
	if len(args) == 3 && args[0] == "add" {
		return reposAdd(args[1], args[2])
	}
	if len(args) == 2 && args[0] == "remove" {
		return reposRemove(args[1])
	}
	fmt.Fprintln(os.Stderr, "Usage: arciv repositoreis show")
	fmt.Fprintln(os.Stderr, "       arciv repositoreis add [repository name] [repository path]")
	fmt.Fprintln(os.Stderr, "       arciv repositoreis remove [repository name]")
	return nil
}

func reposShow() error {
	repos, err := loadRepos()
	if err != nil {
		return err
	}
	for _, repo := range repos {
		fmt.Println(repo)
	}
	return nil
}

func loadRepos() ([]Repository, error) {
	lines, err := loadLines(rootDir + "/.arciv/repositories")
	if err != nil {
		return []Repository{}, err
	}
	repos := []Repository{selfRepo}
	for _, line := range lines {
		idx := strings.Index(line, " ")
		if idx == -1 {
			return []Repository{}, errors.New("Repository path is not registerd in .arciv/repositories")
		}
		name := line[:idx]
		path := line[idx+1:]

		for _, repo := range repos {
			if repo.Name == name {
				return []Repository{}, errors.New("Repositoy name is conflict in .arciv/repositories")
			}
		}
		repos = append(repos, Repository{Name: name, Path: path})
	}
	return repos, nil
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

func reposAdd(name string, path string) error {
	if strings.Index(name, " ") != -1 {
		return errors.New("Repository name must not include space")
	}
	if !strings.HasPrefix(path, "file:///") {
		return errors.New("Repository path must be file:///...")
	}
	repos, err := loadRepos()
	if err != nil {
		return err
	}
	for _, repo := range repos {
		if repo.Name == name {
			return errors.New("The repository name already exists")
		}
	}
	repos = append(repos, Repository{Name: name, Path: path})
	return reposWrite(repos)
}

func reposRemove(name string) error {
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

func reposWrite(repos []Repository) error {
	err := os.Rename(rootDir+"/.arciv/repositories", rootDir+"/.arciv/repositories.org")
	if err != nil {
		return err
	}
	file, err := os.OpenFile(rootDir+"/.arciv/repositories", os.O_WRONLY|os.O_CREATE, 0666)
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

	err = os.Remove(rootDir+"/.arciv/repositories.org")
	if err != nil {
		return err
	}
	return nil
}
