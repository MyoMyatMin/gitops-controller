package sync

import (
	"fmt"
	"path/filepath"

	"github.com/MyoMyatMin/gitops-controller/internal/metrics"

	"github.com/MyoMyatMin/gitops-controller/internal/git"
	"github.com/MyoMyatMin/gitops-controller/internal/k8s"
	"github.com/MyoMyatMin/gitops-controller/internal/log"
	"github.com/MyoMyatMin/gitops-controller/pkg/manifest"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Engine struct {
	gitRepo   *git.Repository
	k8sClient *k8s.Client
	namespace string
	repoPath  string
}

type SyncResult struct {
	CommitSHA string
	Created   []string
	Updated   []string
	Deleted   []string
	Errors    []error
}

func NewEngine(repo *git.Repository, client *k8s.Client, ns, path string) *Engine {
	return &Engine{
		gitRepo:   repo,
		k8sClient: client,
		namespace: ns,
		repoPath:  path,
	}
}

func (e *Engine) Sync() (*SyncResult, error) {

	log.Info("--- Starting Sync ---")

	syncTimer := prometheus.NewTimer(metrics.SyncDuration)
	defer syncTimer.ObserveDuration()

	result := &SyncResult{}

	if err := e.gitRepo.Pull(); err != nil {
		metrics.SyncTotal.WithLabelValues("failure").Inc()
		log.Errorf("error pulling git repo: %v", err)
		return nil, fmt.Errorf("error pulling git repo: %w", err)
	}
	commitSHA, err := e.gitRepo.GetLatestCommit()
	if err != nil {
		metrics.SyncTotal.WithLabelValues("failure").Inc()
		log.Errorf("error getting commit SHA: %v", err)
		return nil, fmt.Errorf("error getting commit SHA: %w", err)
	}
	result.CommitSHA = commitSHA
	log.Infof("Syncing to commit: %s", commitSHA)

	manifestDir := filepath.Join(e.gitRepo.LocalPath, e.repoPath)
	gitManifests, err := ParseManifests(manifestDir)
	if err != nil {
		metrics.SyncTotal.WithLabelValues("failure").Inc()
		log.Errorf("error parsing manifests: %v", err)
		return nil, fmt.Errorf("error parsing manifests: %w", err)
	}

	clusterResources, err := e.k8sClient.ListManagedResources(e.namespace)
	if err != nil {
		metrics.SyncTotal.WithLabelValues("failure").Inc()
		log.Errorf("error listing managed resources: %v", err)
		return nil, fmt.Errorf("error listing managed resources: %w", err)
	}

	toApply, toDelete := e.diff(gitManifests, clusterResources)

	log.Infof("--- Applying %d resources ---", len(toApply))
	for _, m := range toApply {
		m.Object.SetNamespace(e.namespace)
		m.Namespace = e.namespace

		if err := e.k8sClient.Apply(m, false); err != nil {
			result.Errors = append(result.Errors, err)
		} else {
			result.Updated = append(result.Updated, m.Name)
			metrics.ResourceManaged.WithLabelValues("applied", m.Kind).Inc()
		}
	}

	log.Infof("--- Pruning %d resources ---", len(toDelete))
	for _, res := range toDelete {
		m := manifest.Manifest{
			Object:    &res,
			Kind:      res.GetKind(),
			Name:      res.GetName(),
			Namespace: res.GetNamespace(),
		}
		if err := e.k8sClient.Delete(m); err != nil {
			result.Errors = append(result.Errors, err)
		} else {
			result.Deleted = append(result.Deleted, m.Name)
			metrics.ResourceManaged.WithLabelValues("deleted", m.Kind).Inc()
		}
	}

	if len(result.Errors) > 0 {
		metrics.SyncTotal.WithLabelValues("failure").Inc()
	} else {
		metrics.SyncTotal.WithLabelValues("success").Inc()
		metrics.LastSyncTimestamp.SetToCurrentTime()
	}

	log.Info("--- Sync Complete ---")
	return result, nil
}

func (e *Engine) diff(gitManifests []manifest.Manifest, clusterResources []unstructured.Unstructured) (toApply []manifest.Manifest, toDelete []unstructured.Unstructured) {
	toApply = gitManifests

	gitManifestsMap := make(map[string]struct{})
	for _, m := range gitManifests {
		m.Object.SetNamespace(e.namespace)
		key := fmt.Sprintf("%s/%s/%s", m.Kind, m.Object.GetNamespace(), m.Name)
		gitManifestsMap[key] = struct{}{}
	}

	for _, res := range clusterResources {
		key := fmt.Sprintf("%s/%s/%s", res.GetKind(), res.GetNamespace(), res.GetName())
		if _, exists := gitManifestsMap[key]; !exists {
			toDelete = append(toDelete, res)
		}
	}

	return toApply, toDelete
}
