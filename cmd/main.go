package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/MyoMyatMin/gitops-controller/internal/config"

	"github.com/MyoMyatMin/gitops-controller/internal/api"
	"github.com/MyoMyatMin/gitops-controller/internal/git"
	"github.com/MyoMyatMin/gitops-controller/internal/k8s"
	"github.com/MyoMyatMin/gitops-controller/internal/sync"
	"github.com/MyoMyatMin/gitops-controller/pkg/manifest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func main() {
	fmt.Println("GitOps Controller Starting")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	k8sClient, err := k8s.NewClient(cfg.Kubernetes.Kubeconfig)
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	repo := &git.Repository{
		URL:       cfg.Git.URL,
		LocalPath: cfg.Git.LocalPath,
		Branch:    cfg.Git.Branch,
	}
	os.RemoveAll(repo.LocalPath)
	if err := repo.Clone(); err != nil {
		fmt.Printf("Error cloning repository: %v\n", err)
		os.Exit(1)
	}

	targetNamespace := cfg.Kubernetes.Namespace
	targetPath := cfg.Git.Path

	if err := ensureNamespace(k8sClient, targetNamespace); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			fmt.Printf("Error ensuring namespace: %v\n", err)
			os.Exit(1)
		}
	}
	engine := sync.NewEngine(repo, k8sClient, targetNamespace, targetPath)

	poller := sync.NewPoller(engine, cfg.Sync.Interval)
	go poller.Start()

	if cfg.Webhook.Enabled {

		webhookServer := api.NewWebhookServer(engine, cfg.Webhook.Secret)
		go func() {
			if err := webhookServer.Start(cfg.Webhook.Port); err != nil {
				log.Fatalf("Webhook server failed: %v\n", err)
			}
		}()
		fmt.Printf("Webhook server enabled on port %d\n", cfg.Webhook.Port)
	} else {
		fmt.Println("Webhook server is disabled in config.")
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	poller.Stop()

	fmt.Println("Main application shut down gracefully.")

	fmt.Printf("Cleaning up namespace %s...\n", targetNamespace)
	if err := deleteNamespace(k8sClient, targetNamespace); err != nil {
		fmt.Printf("Warning: failed to clean up namespace: %v\n", err)
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
