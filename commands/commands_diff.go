package commands

import (
	"errors"
	"github.com/spf13/cobra"
)

var (
	diffCmd = &cobra.Command{
		Use:   "diff <commit> <commit>",
		Run:   diffCommand,
		Short: "Print difference of commits",
		Long: `Compare commits specified by arguments and print difference.
The command prints changes from the first argument's commit to the second one
Arguments need commit-id, allow a part of commit-id.
For example, if commit-id is '6038d4c5-92e040fe51a920f869a929e3a309e072c7bfe115a1c57b0b472e248f3f09570d',
arguments allow '6038d4', '6038d4c5-92e', '92e', '92e040fe' and so on...
If a part of commit-id points more than 1 commit, an error occurs. `,
		Args: cobra.ExactArgs(2),
	}
)

func diffCommand(cmd *cobra.Command, args []string) {
	if err := diffAction(args[0], args[1]); err != nil {
		Exit(err, 1)
	}
}

func init() {
	RootCmd.AddCommand(diffCmd)
}

func diffAction(commitAlias0, commitAlias1 string) (err error) {
	selfRepo := SelfRepo()
	commit0, err := selfRepo.LoadCommitFromAlias(commitAlias0)
	if err != nil {
		return err
	}
	commit1, err := selfRepo.LoadCommitFromAlias(commitAlias1)
	if err != nil {
		return err
	}
	if commit0.Id == commit1.Id {
		return errors.New("Same commit")
	}
	printDiffs(diffTags(commit0.Tags, commit1.Tags))
	return nil
}

func diffTags(tagsBefore, tagsAfter []Tag) (deleted []Tag, added []Tag) {
	ib, ia := 0, 0
	for ib < len(tagsBefore) && ia < len(tagsAfter) {
		compared := compareTag(tagsBefore[ib], tagsAfter[ia])
		if compared == 0 {
			ib++
			ia++
		} else if compared < 0 {
			deleted = append(deleted, tagsBefore[ib])
			ib++
		} else {
			added = append(added, tagsAfter[ia])
			ia++
		}
	}
	for _, c := range tagsBefore[ib:] {
		deleted = append(deleted, c)
	}
	for _, c := range tagsAfter[ia:] {
		added = append(added, c)
	}
	return deleted, added
}

func printDiffs(deleted, added []Tag) {
	for _, dc := range deleted {
		// same hash
		idx := findTagIndex(added, dc, FIND_HASH|FIND_PATH)
		if idx != -1 {
			message("update: " + dc.Path + ", hash: " + dc.Hash.String() + ", timestamp: \x1b[31m" + timestamp2string(dc.Timestamp) + "\x1b[0m -> \x1b[32m" + timestamp2string(added[idx].Timestamp) + "\x1b[0m")
			added = append(added[:idx], added[idx+1:]...)
			continue
		}
		idx = findTagIndex(added, dc, FIND_HASH)
		if idx != -1 {
			message("rename: \x1b[31m" + dc.Path + "\x1b[0m -> \x1b[32m" + added[idx].Path + "\x1b[0m, hash: " + dc.Hash.String())
			added = append(added[:idx], added[idx+1:]...)
			continue
		}
		// same path, but not same hash
		idx = findTagIndex(added, dc, FIND_PATH)
		if idx != -1 {
			message("rewrite: " + dc.Path + ", hash: \x1b[31m" + dc.Hash.String() + "\x1b[0m -> \x1b[32m" + added[idx].Hash.String() + "\x1b[0m")
			added = append(added[:idx], added[idx+1:]...)
			continue
		}
		// similar tag is not found
		message("\x1b[31mdeleted: " + dc.Path + ", hash: " + dc.Hash.String() + "\x1b[0m")
	}
	// similar tag is not found
	for _, ac := range added {
		message("\x1b[32madded: " + ac.Path + ", hash: " + ac.Hash.String() + "\x1b[0m")
	}
}

func printDiffsSimple(deleted, added []Tag) {
	for _, c := range deleted {
		message("\x1b[31m" + "- " + c.String() + "\x1b[0m")
	}
	for _, c := range added {
		message("\x1b[32m" + "+ " + c.String() + "\x1b[0m")
	}
}
