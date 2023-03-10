package kit_lib

/*
func TestGit2Go(t *testing.T) {
	repo, err := git2go.OpenRepository("/Users/alex/work/kit/Repo2")
	if err != nil {
		panic(err)
	}

	remote, err := repo.Remotes.Lookup("origin")
	if err != nil {
		panic(err)
	}
	if err := remote.Fetch([]string{}, nil, ""); err != nil {
		panic(err)
	}
	_, err = repo.References.Lookup("refs/remotes/origin/main")
	if err != nil {
		panic(err)
	}
	annotatedCommit, err := repo.AnnotatedCommitFromRef(remoteBranch)
	if err != nil {
		panic(err)
	}

	// Do the merge analysis
	mergeHeads := make([]*git2go.AnnotatedCommit, 1)
	mergeHeads[0] = annotatedCommit
	analysis, _, err := repo.MergeAnalysis(mergeHeads)
	if err != nil {
		panic(err)
	}
}

func TestPull(t *testing.T) {
	main.initializeApplication()

	err := main.app.PullWithAbort()
	if err != nil {
		t.Fatal(err)
	}
}
*/

/*func setupSuite(t testing.T) func(t testing.T) {
	log.Println("setup test suite")
	initializeApplication()

	// Return a function to teardown the test
	return func(t testing.T) {
		log.Println("teardown test suite")
	}
}*/
