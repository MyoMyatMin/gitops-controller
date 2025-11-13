package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MyoMyatMin/gitops-controller/internal/log"
	"github.com/MyoMyatMin/gitops-controller/pkg/manifest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func ParseManifests(dirPath string) ([]manifest.Manifest, error) {
	var allManifests []manifest.Manifest

	log.Infof("Starting to parse manifests in: %s", dirPath)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, walkErr error) error {

		if walkErr != nil {
			log.Warnf("Skipping file %s (walk error: %v)", path, walkErr)
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {

			log.Warnf("Skipping file %s (read error: %v)", path, err)
			return nil
		}

		manifests, err := ParseYAML(data)
		if err != nil {

			log.Warnf("Skipping file %s (parse error: %v)", path, err)
			return nil
		}

		allManifests = append(allManifests, manifests...)
		return nil
	})

	if err != nil {
		log.Errorf("error walking directory %s: %v", dirPath, err)
		return nil, fmt.Errorf("error walking directory %s: %w", dirPath, err)
	}

	log.Infof("Finished parsing. Found %d manifests.", len(allManifests))
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
