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
        arciv repository add name:media-stable type:file path:/media/hdd0/arciv-repo-directory
          ... register the new repository, 'mesia-stable' and its root directory is /media/hdd0/arciv-repo-directory
        arciv repository add name:aws-s3-repo type:s3 bucket:s3-bucket-name-hoge region:ap-northeast-1
          ... register the new repository, 'aws-s3-repo' on AWS S3 (ap-northeast-1), s3://s3-bucket-name-hoge
        arciv repository remove media-stable
          ... remove the repository, 'media-stable'
`,
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
	if args[0] == "add" {
		return repositoryActionAdd(args[1:])
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

func repositoryActionAdd(args []string) error {
	for _, elm := range args {
		if strings.Contains(elm, " ") {
			return errors.New("Including space in repository definition is not supported")
		}
	}
	repo, err := strs2repository(args)
	if err != nil {
		return err
	}

	repos, err := loadRepos()
	if err != nil {
		return err
	}
	for _, r := range repos {
		if r.Name == repo.Name {
			return errors.New("The repository name already exists")
		}
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

func strs2repository(elements []string) (Repository, error) {
	var rtype string
	var name string
	var path string
	var region string
	var bucket string
	if len(elements) == 0 {
		return Repository{}, nil
	}
	for _, elm := range elements {
		if elm == "" {
			continue
		}
		if strings.HasPrefix(elm, "type:") {
			if rtype != "" {
				return Repository{}, errors.New("Specifications of repository information is duplicated")
			}
			rtype = elm[len("type:"):]
		} else if strings.HasPrefix(elm, "name:") {
			if name != "" {
				return Repository{}, errors.New("Specifications of repository information is duplicated")
			}
			name = elm[len("name:"):]
		} else if strings.HasPrefix(elm, "path") {
			if path != "" {
				return Repository{}, errors.New("Specifications of repository information is duplicated")
			}
			path = elm[len("path:"):]
		} else if strings.HasPrefix(elm, "region:") {
			if region != "" {
				return Repository{}, errors.New("Specifications of repository information is duplicated")
			}
			region = elm[len("region:"):]
		} else if strings.HasPrefix(elm, "bucket:") {
			if bucket != "" {
				return Repository{}, errors.New("Specifications of repository information is duplicated")
			}
			bucket = elm[len("bucket:"):]
		} else {
			message(elm)
			return Repository{}, errors.New("Repository definition is invalid syntax")
		}
	}
	if name == "" {
		return Repository{}, errors.New("Repository's name is not specified")
	}
	if rtype == "file" {
		if path == "" {
			return Repository{}, errors.New("Repository's type is file, but path is not specified")
		}
		return Repository{Name: name, Location: RepositoryLocationFile{Path: path}}, nil
	}
	if rtype == "s3" {
		if bucket == "" || region == "" {
			return Repository{}, errors.New("Repository's type is file, but bucket or region is not specified")
		}
		return Repository{Name: name, Location: RepositoryLocationS3{RegionName: region, BucketName: bucket}}, nil
	}
	return Repository{}, errors.New("Unknown repository's type")
}

func loadRepos() ([]Repository, error) {
	lines, err := fileOp.loadLines(fileOp.rootDir() + "/.arciv/repositories")
	if err != nil {
		return []Repository{}, err
	}
	repos := []Repository{SelfRepo()}
	for _, line := range lines {
		repo, err := strs2repository(strings.Split(line, " "))
		if err != nil {
			return []Repository{}, err
		}
		for _, r := range repos {
			if r.Name == repo.Name {
				return []Repository{}, errors.New("Repositoy name is conflict in .arciv/repositories")
			}
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
	return fileOp.writeLines(fileOp.rootDir()+"/.arciv/repositories", lines)
}
