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

type Repository struct {
	Name string
	Path string
}

func (repository Repository) String() string {
	return repository.Name + " " + repository.Path
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
	lines, err := loadLines(rootDir() + "/.arciv/repositories")
	if err != nil {
		return []Repository{}, err
	}
	var repos []Repository
	for _, line := range lines {
		idx := strings.Index(line, " ")
		if idx == -1 {
			idx = len(line)
		}
		repos = append(repos, Repository{Name: line[:idx], Path: line[idx+1:]})
	}
	return repos, nil
}

func reposAdd(name string, path string) error {
	if strings.Index(name, " ") != -1 {
		return errors.New("Repository name must not include space")
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
	file, err := os.OpenFile(rootDir()+"/.arciv/repositories", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintln(file, Repository{Name: name, Path: path}.String())
	return nil
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
		fmt.Fprintln(file, repo.String())
	}

	err = os.Remove(root + "/.arciv/repositories.org")
	if err != nil {
		return err
	}
	return nil
}
