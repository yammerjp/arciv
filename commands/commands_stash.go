package commands

import (
	"github.com/spf13/cobra"
	"os"
)

var (
	stashCmd = &cobra.Command{
		Use: "stash",
		Run: stashCommand,
	}
)

func stashCommand(cmd *cobra.Command, args []string) {
	if err := stashAction(); err != nil {
		Exit(err, 1)
	}
}

func stashAction() (err error) {
	photos, err := takePhotosSelfRepo()
	if err != nil {
		return err
	}
	err = stashPhotos(photos)
	if err != nil {
		return err
	}

	message("stashed all files")
	return nil
}

func stashPhotos(photos []Photo) (err error) {
	selfRepo := SelfRepo()
	os.MkdirAll(selfRepo.Path+"/.arciv/blob", 0777)

	// move all files to .arciv/blob
	for _, p := range photos {
		from := selfRepo.Path + "/" + p.Path
		to := selfRepo.Path + "/.arciv/blob/" + p.Hash.String()
		err = os.Rename(from, to)
		if err != nil {
			return err
		}
		message("moved " + from + " -> " + to)
	}

	// remove all directory in root without .arciv
	dirPaths, err := findPathsOfSelfRepo(false, true)
	if err != nil {
		return err
	}
	for i := len(dirPaths) - 1; i >= 0; i-- {
		if os.Remove(dirPaths[i]) == nil {
			message("remove directory, " + dirPaths[i])
		}
	}
	return nil
}

func init() {
	RootCmd.AddCommand(stashCmd)
}
