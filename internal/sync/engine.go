package sync

import (
	"fmt"
	"path/filepath"

	"github.com/MyoMyatMin/gitops-controller/internal/git"
	"github.com/MyoMyatMin/gitops-controller/internal/k8s"
	"github.com/MyoMyatMin/gitops-controller/pkg/manifest"
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
	fmt.Println("--- Starting Sync ---")
	result := &SyncResult{}

	if err := e.gitRepo.Pull(); err != nil {
		return nil, fmt.Errorf("error pulling latest changes: %w", err)
	}
	commitSHA, err := e.gitRepo.GetLatestCommit()
	if err != nil {
		return nil, fmt.Errorf("error getting latest commit SHA: %w", err)
	}

	result.CommitSHA = commitSHA
	fmt.Printf("Syncing to commit: %s\n", commitSHA)

	manifestDir := filepath.Join(e.gitRepo.LocalPath, e.repoPath)
	gitManifests, err := ParseManifests(manifestDir)
	if err != nil {
		return nil, fmt.Errorf("error parsing manifests: %w", err)
	}

	clusterResouces, err := e.k8sClient.ListManagedResources(e.namespace)
	if err != nil {
		return nil, fmt.Errorf("error listing managed resources: %w", err)
	}
	toApply, toDelete := e.diff(gitManifests, clusterResouces)

	fmt.Printf("--- Applying %d manifests ---\n", len(toApply))
	for _, m := range toApply {
		m.Object.SetNamespace(e.namespace)
		if err := e.k8sClient.Apply(m, false); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("error applying %s/%s: %w", m.Namespace, m.Name, err))

		} else {
			result.Updated = append(result.Updated, fmt.Sprintf("%s/%s", m.Namespace, m.Name))
		}
	}

	fmt.Printf("--- Pruning %d resources ---\n", len(toDelete))
	for _, r := range toDelete {
		manifest := manifest.Manifest{
			Object:    &r,
			Name:      r.GetName(),
			Kind:      r.GetKind(),
			Namespace: r.GetNamespace(),
		}
		if err := e.k8sClient.Delete(manifest); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("error pruning %s/%s: %w", manifest.Namespace, manifest.Name, err))
		} else {
			result.Deleted = append(result.Deleted, fmt.Sprintf("%s/%s", manifest.Namespace, manifest.Name))
		}
	}
	fmt.Println("--- Sync Complete ---")
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
