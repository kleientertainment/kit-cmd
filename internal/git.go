package internal

func (repo *git.Repository) Pull() error {
	remote, err := repo.Remotes.Lookup("origin")
	if err != nil {
		return err
	}

	return nil
}
