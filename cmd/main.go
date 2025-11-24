package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/MyoMyatMin/gitops-controller/internal/api"
	"github.com/MyoMyatMin/gitops-controller/internal/config"
	"github.com/MyoMyatMin/gitops-controller/internal/log"

	"github.com/MyoMyatMin/gitops-controller/internal/git"
	"github.com/MyoMyatMin/gitops-controller/internal/k8s"
	"github.com/MyoMyatMin/gitops-controller/internal/sync"
	"github.com/MyoMyatMin/gitops-controller/pkg/manifest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func main() {

	log.Init()
	log.Info("GitOps Controller starting")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	k8sClient, err := k8s.NewClient(cfg.Kubernetes.Kubeconfig)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	var engines []*sync.Engine
	var pollers []*sync.Poller

	for _, repoCfg := range cfg.Repositories {
		log.Infof("Initializing repository: %s", repoCfg.Name)

		localPath := filepath.Join("/tmp/gitops-repos", repoCfg.Name)

		repo := &git.Repository{
			URL:       repoCfg.URL,
			Branch:    repoCfg.Branch,
			LocalPath: localPath,
		}

		os.RemoveAll(repo.LocalPath)
		if err := repo.Clone(); err != nil {
			log.Errorf("Failed to clone repo %s: %v", repoCfg.Name, err)
		}

		if err := ensureNamespace(k8sClient, repoCfg.Namespace); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				log.Errorf("Error ensuring namespace %s: %v", repoCfg.Namespace, err)
			}
		}

		engine := sync.NewEngine(repo, k8sClient, repoCfg.Namespace, repoCfg.Path)
		engines = append(engines, engine)

		poller := sync.NewPoller(engine, repoCfg.Interval)
		pollers = append(pollers, poller)

		go poller.Start()
	}

	if cfg.Webhook.Enabled {
		webhookServer := api.NewWebhookServer(engines, cfg.Webhook.Secret)
		go func() {
			if err := webhookServer.Start(cfg.Webhook.Port); err != nil {
				log.Fatalf("Webhook server failed: %v", err)
			}
		}()
		log.Infof("Webhook server enabled on port %d", cfg.Webhook.Port)
	}
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Info("Shutting down...")

	// Stop all pollers
	for _, p := range pollers {
		p.Stop()
	}

	log.Info("Main application shut down gracefully.")
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
