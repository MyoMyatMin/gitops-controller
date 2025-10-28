package main

import (
	"fmt"
	"os"

	"github.com/MyoMyatMin/gitops-controller/internal/git"
	"github.com/MyoMyatMin/gitops-controller/internal/k8s"
)

func main() {
	fmt.Println("GitOps Controller Starting")
	k8sClient, err := k8s.NewClient("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	err = k8sClient.ListNamespaces()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing namespaces: %v\n", err)
		os.Exit(1)
	}

	testRepoURL := "https://github.com/kubernetes/examples.git"
	testRepoBranch := "master"
	localPath := "/tmp/gitops-test-repo"

	fmt.Printf("Cleaning up old test repo at %s...\n", localPath)
	os.RemoveAll(localPath)

	repo := &git.Repository{
		URL:       testRepoURL,
		LocalPath: localPath,
		Branch:    testRepoBranch,
	}

	fmt.Println("\n--- Testing Clone ---")
	if err := repo.Clone(); err != nil {
		fmt.Printf("Error cloning repository: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n--- Testing GetLatestCommit ---")
	sha1, err := repo.GetLatestCommit()
	if err != nil {
		fmt.Printf("Error getting commit: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n--- Testing Pull ---")
	if err := repo.Pull(); err != nil {
		fmt.Printf("Error pulling repository: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n--- Testing GetLatestCommit (after pull) ---")
	sha2, err := repo.GetLatestCommit()
	if err != nil {
		fmt.Printf("Error getting commit: %v\n", err)
		os.Exit(1)
	}

	if sha1 != sha2 {
		fmt.Println("WARN: Commit SHA changed between clone and pull, which is fine.")
	}

	fmt.Printf("\nTest successful. Cloned repo at %s. Exiting.\n", localPath)
	os.Exit(0)

}
