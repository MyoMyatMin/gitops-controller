package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MyoMyatMin/gitops-controller/pkg/manifest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (c *Client) Apply(manifest manifest.Manifest, dryRun bool) error {

	obj := manifest.Object
	if obj == nil {
		return fmt.Errorf("manifest object is nil for %s", manifest.Name)
	}

	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	// --- FIX 1 ---
	// Use the constant, not a hardcoded string
	labels[ManagedByLabel] = FieldManager
	obj.SetLabels(labels)

	resourceInterface, err := c.getResourceInterface(manifest)
	if err != nil {
		return err
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal object %s/%s: %v", manifest.Namespace, manifest.Name, err)
	}

	patchOptions := metav1.PatchOptions{
		FieldManager: FieldManager, // Use constant here too
		Force:        boolPtr(true),
	}

	if dryRun {
		patchOptions.DryRun = []string{metav1.DryRunAll}
		fmt.Printf("Dry-Run Applying: Kind=%s, Name=%s, Namespace=%s\n",
			obj.GetKind(), obj.GetName(), obj.GetNamespace())
	} else {
		fmt.Printf("Applying: Kind=%s, Name=%s, Namespace=%s\n",
			obj.GetKind(), obj.GetName(), obj.GetNamespace())
	}

	// --- FIX 2 ---
	// This line was a duplicate and printed twice. Remove it.
	// fmt.Printf("Applying: Kind=%s Name=%s Namespace=%s\n", ...)

	_, err = resourceInterface.Patch(
		context.TODO(),
		obj.GetName(),
		types.ApplyPatchType,
		data,
		patchOptions,
	)
	if err != nil {
		return fmt.Errorf("failed to apply object %s/%s: %v", manifest.Namespace, manifest.Name, err)
	}

	return nil
}

func boolPtr(b bool) *bool {
	return &b
}
