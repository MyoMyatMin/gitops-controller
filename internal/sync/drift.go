package sync

import (
	"fmt"
	"reflect"

	"github.com/MyoMyatMin/gitops-controller/pkg/manifest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func DetectDrift(gitManifests []manifest.Manifest, clusterResources []unstructured.Unstructured) (bool, []string) {
	var driftReasons []string
	hasDrift := false

	clusterMap := make(map[string]unstructured.Unstructured)
	for _, res := range clusterResources {
		key := fmt.Sprintf("%s/%s/%s", res.GetKind(), res.GetNamespace(), res.GetName())
		clusterMap[key] = res
	}

	for _, gitRes := range gitManifests {
		key := fmt.Sprintf("%s/%s/%s", gitRes.Kind, gitRes.Object.GetNamespace(), gitRes.Name)
		clusterRes, exists := clusterMap[key]
		if !exists {
			hasDrift = true
			driftReasons = append(driftReasons, fmt.Sprintf("Resource missing in cluster: %s", key))
			continue
		}

		diffs := compareObjects(gitRes.Object.Object, clusterRes.Object)
		if len(diffs) > 0 {
			hasDrift = true
			for _, diff := range diffs {
				driftReasons = append(driftReasons, fmt.Sprintf("%s: %s", key, diff))
			}
		}
	}
	return hasDrift, driftReasons
}

func compareObjects(expected, actual map[string]interface{}) []string {

	var diffs []string

	diffs = append(diffs, checkMapSubset(expected, actual, "")...)
	return diffs
}

func checkMapSubset(expected, actual map[string]interface{}, path string) []string {
	var diffs []string

	for key, expectedVal := range expected {
		if path == "" && (key == "apiVersion" || key == "status" || key == "kind") {
			continue
		}
		if path == ".metadata" && (key == "uid" || key == "resourceVersion" || key == "creationTimestamp" || key == "generation" || key == "managedFields") {
			continue
		}

		actualVal, exists := actual[key]
		if !exists {
			diffs = append(diffs, fmt.Sprintf("Missing field at %s.%s", path, key))
			continue
		}

		if expectedMap, ok := expectedVal.(map[string]interface{}); ok {
			if actualMap, ok := actualVal.(map[string]interface{}); ok {
				diffs = append(diffs, checkMapSubset(expectedMap, actualMap, path+"."+key)...)
				continue
			}
		}
		if !reflect.DeepEqual(expectedVal, actualVal) {
			diffs = append(diffs, fmt.Sprintf("Drift at %s.%s: Git='%v', Cluster='%v'", path, key, expectedVal, actualVal))
		}
	}
	return diffs
}
