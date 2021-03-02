package commands

import (
	"errors"
	"github.com/spf13/cobra"
	"strings"
)

var (
	repositoryCmd = &cobra.Command{
		Use:   "repository ( | add <name> <path> | remove <name>)",
		Run:   repositoryCommand,
		Short: "Show, add or remove repositories",
		Long: `Show, add or remove repositories.
On excute 'arciv repository', the command shows repositories registerd the current repository.
On excite 'arciv repository add', the command registers a new repository to the current repository.
On excute 'arciv repository remove', the command removes a already registerd repository from the current repository.

Example:
        arciv repository add media-stable file:///media/hdd0/arciv-repo-directory
          ... register the new repository, 'mesia-stable' and its root directory is /media/hdd0/arciv-repo-directory
        arciv repository remove media-stable
          ... remove the repository, 'media-stable'
`,
		Args: cobra.RangeArgs(0, 3),
	}
)

func repositoryCommand(cmd *cobra.Command, args []string) {
	if err := repositoryAction(args); err != nil {
		Exit(err, 1)
	}
}

func init() {
	RootCmd.AddCommand(repositoryCmd)
}

func repositoryAction(args []string) (err error) {
	if len(args) == 0 {
		return repositoryActionShow()
	}
	if len(args) == 3 && args[0] == "add" {
		return repositoryActionAdd(args[1], args[2])
	}
	if len(args) == 2 && args[0] == "remove" {
		return repositoryActionRemove(args[1])
	}
	message("Usage: arciv repository")
	message("       arciv repository add [repository name] [repository path]")
	message("       arciv repository remove [repository name]")
	return nil
}

func repositoryActionShow() error {
	repos, err := loadRepos()
	if err != nil {
		return err
	}
	for _, repo := range repos {
		messageStdin(repo.String())
	}
	return nil
}

func repositoryActionAdd(name string, url string) error {
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
	err = repo.Init()
	if err != nil {
		return err
	}
	return writeRepos(append(repos, repo))
}

func repositoryActionRemove(name string) error {
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
		return writeRepos(repos)
	}
	return errors.New("The repository is not found")
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

func writeRepos(repos []Repository) error {
	var lines []string
	for _, repo := range repos {
		if repo.Name == "self" {
			continue
		}
		lines = append(lines, repo.String())
	}
	return writeLines(rootDir()+"/.arciv/repositories", lines)
}
