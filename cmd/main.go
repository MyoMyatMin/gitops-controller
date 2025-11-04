package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MyoMyatMin/gitops-controller/internal/git"
	"github.com/MyoMyatMin/gitops-controller/internal/k8s"
	"github.com/MyoMyatMin/gitops-controller/internal/sync"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func main() {
	fmt.Println("GitOps Controller Starting")

	fmt.Println("\n--- Connecting to Kubernetes ---")
	k8sClient, err := k8s.NewClient("")
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	testRepoURL := "https://github.com/argoproj/argocd-example-apps.git"
	localPath := "/tmp/gitops-test-repo"
	fmt.Printf("\nCleaning up old test repo at %s...\n", localPath)
	os.RemoveAll(localPath)
	repo := &git.Repository{
		URL:       testRepoURL,
		LocalPath: localPath,
		Branch:    "master",
	}
	if err := repo.Clone(); err != nil {
		fmt.Printf("Error cloning repository: %v\n", err)
		os.Exit(1)
	}

	manifestPath := filepath.Join(localPath, "guestbook")
	fmt.Printf("\n--- Parsing manifests in: %s ---\n", manifestPath)
	manifests, err := sync.ParseManifests(manifestPath)
	if err != nil {
		fmt.Printf("Error parsing manifests: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully parsed %d manifests.\n", len(manifests))

	fmt.Println("\n--- Testing Apply, Get, Delete ---")

	targetNamespace := "guestbook-test" // Use a new name for safety
	nsManifest := buildNamespaceManifest(targetNamespace)

	fmt.Println("\n--- 1. Testing Dry-Run Apply ---")
	if err := k8sClient.Apply(nsManifest, true); err != nil {
		fmt.Printf("Error on dry-run apply: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Dry-Run apply successful. (Nothing was created)")

	fmt.Println("\n--- 2. Testing Real Apply ---")
	if err := k8sClient.Apply(nsManifest, false); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			fmt.Printf("Error applying namespace (aborting): %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Namespace already exists, continuing.")
	} else {
		fmt.Println("Namespace applied. Waiting 5s for it to be ready...")
		time.Sleep(5 * time.Second)
	}

	for _, m := range manifests {
		m.Object.SetNamespace(targetNamespace)
		m.Namespace = targetNamespace
		if err := k8sClient.Apply(m, false); err != nil {
			fmt.Printf("Error applying manifest %s %s: %v\n", m.Kind, m.Name, err)
		}
	}

	// 3. Test Get
	fmt.Println("\n--- 3. Testing Get ---")
	if len(manifests) > 0 {
		testManifest := manifests[0]
		testManifest.Object.SetNamespace(targetNamespace)
		testManifest.Namespace = targetNamespace

		obj, err := k8sClient.Get(testManifest)
		if err != nil {
			fmt.Printf("Error getting manifest: %v\n", err)
		} else {
			fmt.Printf("Successfully got manifest: %s, uid: %s\n", obj.GetName(), obj.GetUID())
		}
	}

	fmt.Println("\n--- 4. Testing Delete (Cleanup) ---")
	// Delete manifests in reverse (services before deployments, etc.)
	for i := len(manifests) - 1; i >= 0; i-- {
		m := manifests[i]
		m.Object.SetNamespace(targetNamespace)
		m.Namespace = targetNamespace
		if err := k8sClient.Delete(m); err != nil {
			fmt.Printf("Error deleting manifest %s %s: %v\n", m.Kind, m.Name, err)
		}
	}

	if err := k8sClient.Delete(nsManifest); err != nil {
		fmt.Printf("Error deleting namespace: %v\n", err)
	}
	fmt.Println("Cleanup complete.")

	fmt.Println("\nTest successful. Exiting.")
	os.Exit(0)
}

func buildNamespaceManifest(name string) sync.Manifest {
	return sync.Manifest{
		Kind: "Namespace",
		Name: name,
		Object: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": name,
				},
			},
		},
	}
}
