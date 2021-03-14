package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type RestoreRequest struct {
	Repository Repository
	ValidDays  int32
	Commit     Commit
	Blobs      []string
}

func (r RestoreRequest) String() (str string) {
	for _, line := range r.Strings() {
		str += line + "\n"
	}
	return str
}

func (r RestoreRequest) Strings() []string {
	strs := []string{
		"#arciv-restore-request",
		"#valid-days:" + fmt.Sprintf("%d", r.ValidDays),
		"#repo:" + r.Repository.String(),
		"#commit:" + r.Commit.Id,
	}
	return append(strs, r.Blobs...)
}

func strs2restoreRequest(lines []string) (RestoreRequest, error) {
	if len(lines) <= 4 {
		return RestoreRequest{}, errors.New("The number of lines is small")
	}

	// line 0
	if lines[0] != "#arciv-restore-request" {
		return RestoreRequest{}, errors.New("The line 0 is invalid syntax")
	}

	// line 1
	if !strings.HasPrefix(lines[1], "#valid-days:") {
		return RestoreRequest{}, errors.New("The line 1 is invalid syntax")
	}
	validDaysInt, err := strconv.ParseInt(lines[1][len("#valid-days:"):], 10, 32)
	if err != nil {
		return RestoreRequest{}, err
	}
	validDays := int32(validDaysInt)

	// line 2
	if !strings.HasPrefix(lines[2], "#repo:") {
		return RestoreRequest{}, errors.New("The line 2 is invalid syntax")
	}
	repo, err := strs2repository(strings.Split(lines[2][len("#repo:"):], " "))
	if err != nil {
		return RestoreRequest{}, errors.New("The line 2 is invalid syntax")
	}

	// line 3
	if !strings.HasPrefix(lines[3], "#commit:") || len(lines[3]) != len("#commit:")+64 {
		return RestoreRequest{}, errors.New("The line 3 is invalid syntax")
	}
	commit, err := repo.LoadCommit(lines[3][len("#commit:"):])
	if err != nil {
		return RestoreRequest{}, err
	}

	// other lines
	for _, line := range lines[4:] {
		if len(line) != 64 {
			return RestoreRequest{}, errors.New("Invalid syntax line is found")
		}
	}
	return RestoreRequest{
		Repository: repo,
		ValidDays:  validDays,
		Commit:     commit,
		Blobs:      lines[4:],
	}, nil
}
