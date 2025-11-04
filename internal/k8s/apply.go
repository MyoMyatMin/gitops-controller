package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MyoMyatMin/gitops-controller/internal/sync"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (c *Client) Apply(manifest sync.Manifest, dryRun bool) error {

	obj := manifest.Object
	if obj == nil {
		return fmt.Errorf("manifest object is nil for %s", manifest.Name)
	}
	resourceInterface, err := c.getResourceInterface(manifest)
	if err != nil {
		return err
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal object %s/%s: %v", manifest.Namespace, manifest.Name, err)
	}

	fieldManager := "gitops-controller"

	patchOptions := metav1.PatchOptions{
		FieldManager: fieldManager,
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
	fmt.Printf("Applying: Kind=%s Name=%s Namespace=%s\n",
		obj.GetKind(), obj.GetName(), obj.GetNamespace())

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
