package main

import (
	"fmt"
	"os"
	"path/filepath"

	// Use your module path
	"github.com/MyoMyatMin/gitops-controller/internal/git"
	"github.com/MyoMyatMin/gitops-controller/internal/sync"
)

func main() {
	fmt.Println("GitOps Controller Starting")

	testRepoURL := "https://github.com/argoproj/argocd-example-apps.git"
	testRepoBranch := "master"
	localPath := "/tmp/gitops-test-repo"

	fmt.Printf("Cleaning up old test repo at %s...\n", localPath)
	os.RemoveAll(localPath)

	repo := &git.Repository{
		URL:       testRepoURL,
		LocalPath: localPath,
		Branch:    testRepoBranch,
	}

	if err := repo.Clone(); err != nil {
		fmt.Printf("Error cloning repository: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("\n--- Repository cloned successfully ---")

	fmt.Printf("\n--- Listing contents of %s ---\n", localPath)
	entries, err := os.ReadDir(localPath)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
	} else {
		fmt.Println("Directories in repo:")
		for _, entry := range entries {
			if entry.IsDir() && entry.Name()[0] != '.' {
				fmt.Printf("  - %s\n", entry.Name())
			}
		}
	}
	manifestPath := filepath.Join(localPath, "guestbook")

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		fmt.Printf("\nError: Manifest path does not exist: %s\n", manifestPath)
		fmt.Println("Please check the available directories listed above.")
		os.Exit(1)
	}

	fmt.Printf("\n--- Parsing manifests in: %s ---\n", manifestPath)

	manifests, err := sync.ParseManifests(manifestPath)
	if err != nil {
		fmt.Printf("Error parsing manifests: %v\n", err)
		os.Exit(1)
	}

	if len(manifests) == 0 {
		fmt.Println("No manifests found.")
	} else {
		fmt.Printf("\nSuccessfully parsed %d manifests:\n", len(manifests))
		for _, m := range manifests {
			fmt.Printf("- File: %s, Kind: %s, Name: %s\n",
				filepath.Base(m.FilePath), m.Kind, m.Name)
		}
	}

	fmt.Println(" Test successful. Exiting.")
	os.Exit(0)
}
