package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MyoMyatMin/gitops-controller/pkg/manifest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func ParseManifests(dirPath string) ([]manifest.Manifest, error) {
	var allManifests []manifest.Manifest

	fmt.Printf("Starting to parse manifests in: %s\n", dirPath)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Warning: skipping file %s (read error: %v)\n", path, err)
			return nil
		}

		manifests, err := ParseYAML(data)
		if err != nil {
			fmt.Printf("Warning: skipping file %s (parse error: %v)\n", path, err)
			return nil
		}

		for i := range manifests {
			manifests[i].FilePath = path
		}
		allManifests = append(allManifests, manifests...)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %w", dirPath, err)
	}

	fmt.Printf("Finished parsing. Found %d manifests.\n", len(allManifests))
	return allManifests, nil
}

func ParseYAML(data []byte) ([]manifest.Manifest, error) {
	var manifests []manifest.Manifest

	docs := strings.Split(string(data), "\n---\n")

	for _, docData := range docs {

		docData = strings.TrimSpace(docData)
		if docData == "" {
			continue
		}

		obj := &unstructured.Unstructured{Object: make(map[string]interface{})}
		if err := yaml.Unmarshal([]byte(docData), &obj.Object); err != nil {
			return nil, fmt.Errorf("error unmarshaling YAML: %w", err)
		}

		if obj.GetKind() == "" || obj.GetAPIVersion() == "" || obj.GetName() == "" {
			continue
		}

		manifests = append(manifests, manifest.Manifest{
			Kind:      obj.GetKind(),
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
			Object:    obj,
		})
	}

	return manifests, nil
}
