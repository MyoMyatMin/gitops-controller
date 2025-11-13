package git

import (
	"fmt"
	"os"

	"github.com/MyoMyatMin/gitops-controller/internal/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sirupsen/logrus"
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
				log.Infof("Repository already cloned at %s", r.LocalPath)
				return nil
			}
		}
	}

	log.Infof("Cloning repository %s to %s...", r.URL, r.LocalPath)

	_, err := git.PlainClone(r.LocalPath, false, &git.CloneOptions{
		URL:           r.URL,
		ReferenceName: plumbing.NewBranchReferenceName(r.Branch),
		Progress:      log.Logger.WriterLevel(logrus.DebugLevel),
		SingleBranch:  true,
	})

	if err != nil {
		if err == git.ErrRepositoryAlreadyExists {
			log.Warn("Clone failed: Repository already exists.")
			return nil
		}
		log.Errorf("error cloning repository: %v", err)
		return fmt.Errorf("error cloning repository: %w", err)
	}

	log.Info("Repository cloned successfully.")
	return nil
}

func (r *Repository) Pull() error {
	log.Info("Pulling latest changes...")
	repo, err := git.PlainOpen(r.LocalPath)
	if err != nil {
		log.Errorf("error opening repository at %s: %v", r.LocalPath, err)
		return fmt.Errorf("error opening repository at %s: %w", r.LocalPath, err)
	}

	w, err := repo.Worktree()
	if err != nil {
		log.Errorf("error getting worktree: %v", err)
		return fmt.Errorf("error getting worktree: %w", err)
	}

	err = w.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(r.Branch),
		Progress:      log.Logger.WriterLevel(logrus.DebugLevel),
	})

	if err != nil && err != git.NoErrAlreadyUpToDate {
		log.Errorf("error pulling changes: %v", err)
		return fmt.Errorf("error pulling changes: %w", err)
	}

	log.Info("Pull successful. Repository is up-to-date.")
	return nil
}

func (r *Repository) GetLatestCommit() (string, error) {
	repo, err := git.PlainOpen(r.LocalPath)
	if err != nil {
		log.Errorf("error opening repository at %s: %v", r.LocalPath, err)
		return "", fmt.Errorf("error opening repository at %s: %w", r.LocalPath, err)
	}

	headRef, err := repo.Head()
	if err != nil {
		log.Errorf("error getting HEAD: %v", err)
		return "", fmt.Errorf("error getting HEAD: %w", err)
	}

	commitSHA := headRef.Hash().String()
	log.Infof("Latest commit SHA: %s", commitSHA)
	return commitSHA, nil
}
