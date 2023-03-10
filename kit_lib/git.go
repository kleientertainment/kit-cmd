package kit_lib

import (
	"errors"
	"github.com/libgit2/git2go/v34"
)

var ErrUpToDate = errors.New("up to date")
var ErrUnborn = errors.New("unborn repo")
var ErrNoMergePossible = errors.New("no merge possible")

func Pull(repo *git.Repository) error {
	// locate remote
	remote, err := repo.Remotes.Lookup("origin")
	if err != nil {
		return err
	}
	// fetch changes from remote
	if err = remote.Fetch([]string{}, nil, ""); err != nil {
		return err
	}
	// get corresponding remote reference
	remoteBranch, err := repo.References.Lookup("refs/remotes/origin/main")
	if err != nil {
		return err
	}
	// perform merge analysis
	annotatedCommit, err := repo.AnnotatedCommitFromRef(remoteBranch)
	if err != nil {
		return err
	}
	mergeHeads := make([]*git.AnnotatedCommit, 1)
	mergeHeads[0] = annotatedCommit
	analysis, _, err := repo.MergeAnalysis(mergeHeads) // TODO: MergePreference
	if err != nil {
		return err
	}
	// check value of analysis to see what merge is available
	if analysis&git.MergeAnalysisUpToDate == git.MergeAnalysisUpToDate {
		return ErrUpToDate
	} else if analysis&git.MergeAnalysisUnborn == git.MergeAnalysisUnborn {
		return ErrUnborn
	} else if (analysis&git.MergeAnalysisFastForward == git.MergeAnalysisFastForward) && (analysis&git.MergeAnalysisNormal == git.MergeAnalysisNormal) {
		// do fast forward
	} else if analysis&git.MergeAnalysisNormal == git.MergeAnalysisNormal {
		// do normal merge
	} else {
		return ErrNoMergePossible
	}

	return nil
}
