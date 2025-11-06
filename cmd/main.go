package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/MyoMyatMin/gitops-controller/internal/git"
	"github.com/MyoMyatMin/gitops-controller/internal/k8s"
	"github.com/MyoMyatMin/gitops-controller/internal/sync"
	"github.com/MyoMyatMin/gitops-controller/pkg/manifest"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func main() {
	fmt.Println("GitOps Controller Starting")

	k8sClient, err := k8s.NewClient("")
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	repo := &git.Repository{
		URL:       "https://github.com/argoproj/argocd-example-apps.git",
		LocalPath: "/tmp/gitops-test-repo",
		Branch:    "master",
	}

	os.RemoveAll(repo.LocalPath)
	if err := repo.Clone(); err != nil {
		fmt.Printf("Error cloning repository: %v\n", err)
		os.Exit(1)
	}

	targetNamespace := "guestbook-prod"
	targetPath := "guestbook"

	if err := ensureNamespace(k8sClient, targetNamespace); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			fmt.Printf("Error ensuring namespace: %v\n", err)
			os.Exit(1)
		}
	}

	engine := sync.NewEngine(repo, k8sClient, targetNamespace, targetPath)

	fmt.Println("\n\n--- RUNNING FIRST SYNC (CREATE) ---")
	result1, err := engine.Sync()
	if err != nil {
		fmt.Printf("Error on first sync: %v\n", err)
		os.Exit(1)
	}
	printSyncResult(result1)

	fmt.Println("\n\n--- RUNNING SECOND SYNC (NO-OP) ---")
	result2, err := engine.Sync()
	if err != nil {
		fmt.Printf("Error on second sync: %v\n", err)
		os.Exit(1)
	}
	printSyncResult(result2)

	fmt.Println("\n--- Test complete! ---")
	fmt.Printf("Run 'kubectl -n %s get all' to see the app.\n", targetNamespace)
	fmt.Printf("Run 'kubectl delete ns %s' to clean up.\n", targetNamespace)
}

func ensureNamespace(c *k8s.Client, name string) error {
	nsManifest := manifest.Manifest{
		Object: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata":   map[string]interface{}{"name": name},
			},
		},
	}
	return c.Apply(nsManifest, false)
}

func printSyncResult(r *sync.SyncResult) {
	fmt.Printf("Sync to commit %s complete.\n", r.CommitSHA)
	fmt.Printf("- Updated: %d\n", len(r.Updated))
	fmt.Printf("- Deleted: %d\n", len(r.Deleted))
	if len(r.Errors) > 0 {
		fmt.Printf("- Errors: %d\n", len(r.Errors))
		for _, e := range r.Errors {
			fmt.Printf("  - %v\n", e)
		}
	}
}
