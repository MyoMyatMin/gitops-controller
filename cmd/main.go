package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

	targetNamespace := "guestbook-poller"
	targetPath := "guestbook"

	if err := ensureNamespace(k8sClient, targetNamespace); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			fmt.Printf("Error ensuring namespace: %v\n", err)
			os.Exit(1)
		}
	}
	engine := sync.NewEngine(repo, k8sClient, targetNamespace, targetPath)

	pollInterval := 10 * time.Second
	poller := sync.NewPoller(engine, pollInterval)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go poller.Start()

	<-sigCh

	poller.Stop()

	fmt.Println("Main application shut down gracefully.")

	fmt.Printf("Cleaning up namespace %s...\n", targetNamespace)

	if err := deleteNamespace(k8sClient, targetNamespace); err != nil {
		fmt.Println("Warning: failed to clean up namespace.")
	}
}

func ensureNamespace(c *k8s.Client, name string) error {
	nsManifest := manifest.Manifest{
		Kind: "Namespace",
		Name: name,
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

func deleteNamespace(c *k8s.Client, name string) error {
	nsManifest := manifest.Manifest{
		Kind: "Namespace",
		Name: name,
		Object: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata":   map[string]interface{}{"name": name},
			},
		},
	}
	return c.Delete(nsManifest)
}
