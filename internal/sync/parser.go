package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Manifest struct {
	FilePath  string
	Kind      string
	Name      string
	Namespace string
	Object    *unstructured.Unstructured
}

func ParseYAML(data []byte) ([]Manifest, error) {
	var manifests []Manifest

	docs := strings.Split(string(data), "\n---\n") // why --- here? bcuz multi-doc YAML separator

	for _, docData := range docs {
		docData = strings.TrimSpace(docData)
		if docData == "" {
			continue
		}

		obj := &unstructured.Unstructured{Object: make(map[string]interface{})}

		if err := yaml.Unmarshal([]byte(docData), &obj.Object); err != nil {
			return nil, err
		}

		if obj.GetKind() == "" || obj.GetAPIVersion() == "" || obj.GetName() == "" {
			continue
		}

		manifests = append(manifests, Manifest{
			Kind:      obj.GetKind(),
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
			Object:    obj,
		})

	}

	return manifests, nil
}

func ParseManifests(dirPath string) ([]Manifest, error) {
	var allManifests []Manifest

	fmt.Printf("Parsing manifests in directory: %s\n", dirPath)

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
			return err
		}

		manifests, err := ParseYAML(data)
		if err != nil {
			fmt.Printf("Warning: skipping file %s (read error: %v)\n", path, err)
			return nil
		}

		for i := range manifests {
			manifests[i].FilePath = path
		}

		allManifests = append(allManifests, manifests...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return allManifests, nil
}
