package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MyoMyatMin/gitops-controller/internal/log"
	"github.com/MyoMyatMin/gitops-controller/pkg/manifest"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (c *Client) Apply(manifest manifest.Manifest, dryRun bool) error {
	obj := manifest.Object
	if obj == nil {
		log.Errorf("manifest object is nil for %s", manifest.Name)
		return fmt.Errorf("manifest object is nil for %s", manifest.Name)
	}

	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[ManagedByLabel] = FieldManager
	obj.SetLabels(labels)

	resourceInterface, err := c.getResourceInterface(manifest)
	if err != nil {
		return err
	}

	data, err := json.Marshal(obj)
	if err != nil {
		log.Errorf("error marshaling object to JSON for %s: %v", obj.GetName(), err)
		return fmt.Errorf("error marshaling object to JSON for %s: %w", obj.GetName(), err)
	}

	patchOptions := metav1.PatchOptions{
		FieldManager: FieldManager,
		Force:        boolPtr(true),
	}

	logFields := logrus.Fields{
		"kind":      obj.GetKind(),
		"name":      obj.GetName(),
		"namespace": obj.GetNamespace(),
	}

	if dryRun {
		patchOptions.DryRun = []string{metav1.DryRunAll}
		log.WithFields(logFields).Info("Dry-Run Applying resource")
	} else {
		log.WithFields(logFields).Info("Applying resource")
	}

	_, err = resourceInterface.Patch(
		context.TODO(),
		obj.GetName(),
		types.ApplyPatchType,
		data,
		patchOptions,
	)

	if err != nil {
		log.WithFields(logFields).Errorf("Error applying resource: %v", err)
		return err
	}

	return nil
}

func boolPtr(b bool) *bool {
	return &b
}
