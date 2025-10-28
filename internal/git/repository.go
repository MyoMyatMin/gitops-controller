package git

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Repository struct {
	URL       string
	LocalPath string
	Branch    string
}

func (r *Repository) Clone() error {

	if _, err := os.Stat(r.LocalPath); !os.IsNotExist(err) {

		repo, err := git.PlainOpen(r.LocalPath)
		if err == nil {

			remote, err := repo.Remote("origin")
			if err == nil && len(remote.Config().URLs) > 0 && remote.Config().URLs[0] == r.URL {
				fmt.Printf("Repository already cloned at %s\n", r.LocalPath)
				return nil
			}
		}

		fmt.Printf("Directory %s exists but is not the correct repository.\n", r.LocalPath)
	}

	fmt.Printf("Cloning repository %s to %s...\n", r.URL, r.LocalPath)

	_, err := git.PlainClone(r.LocalPath, false, &git.CloneOptions{
		URL:           r.URL,
		ReferenceName: plumbing.NewBranchReferenceName(r.Branch),
		Progress:      os.Stdout,
		SingleBranch:  true,
	})

	if err != nil {

		if err == git.ErrRepositoryAlreadyExists {
			fmt.Println("Clone failed: Repository already exists.")
			return nil
		}
		return fmt.Errorf("error cloning repository: %w", err)
	}

	fmt.Println("Repository cloned successfully.")
	return nil
}

func (r *Repository) Pull() error {
	fmt.Printf("Pulling latest changes for %s...\n", r.LocalPath)

	repo, err := git.PlainOpen(r.LocalPath)
	if err != nil {
		return fmt.Errorf("error opening repository at %s: %w", r.LocalPath, err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("error getting worktree: %w", err)
	}

	fmt.Println("Pulling latest changes...")
	err = w.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(r.Branch),
		Progress:      os.Stdout,
	})

	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("error pulling changes: %w", err)
	}

	fmt.Println("Pull successful. Repository is up-to-date.")
	return nil
}

func (r *Repository) GetLatestCommit() (string, error) {
	repo, err := git.PlainOpen(r.LocalPath)
	if err != nil {
		return "", fmt.Errorf("error opening repository at %s: %w", r.LocalPath, err)
	}

	headRef, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("error getting HEAD: %w", err)
	}

	commitSHA := headRef.Hash().String()
	fmt.Printf("Latest commit SHA: %s\n", commitSHA)
	return commitSHA, nil
}
