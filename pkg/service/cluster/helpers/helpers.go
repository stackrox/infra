package helpers

import (
	"encoding/json"
	"strings"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/stackrox/infra/pkg/logging"
)

var (
	log = logging.CreateProductionLogger()
)

// TODO: what does this do and is it still required? Is this related to an Argo migration or GCP migration?
// https://github.com/stackrox/infra/pull/695
func HandleArtifactMigration(workflow v1alpha1.Workflow, artifact v1alpha1.Artifact) (string, string) {
	var bucket string
	var key string

	if workflow.Status.ArtifactRepositoryRef != nil &&
		workflow.Status.ArtifactRepositoryRef.ArtifactRepository.GCS != nil &&
		workflow.Status.ArtifactRepositoryRef.ArtifactRepository.GCS.Bucket != "" {
		bucket = workflow.Status.ArtifactRepositoryRef.ArtifactRepository.GCS.Bucket
	} else if artifact.GCS != nil && artifact.GCS.Bucket != "" {
		bucket = artifact.GCS.Bucket
	}

	if artifact.GCS != nil && artifact.GCS.Key != "" {
		key = artifact.GCS.Key
	}

	if bucket == "" || key == "" {
		log.Log(logging.WARN, "cannot figure out bucket for artifact, possibly an upgrade issue, not fatal",
			"workflow-name", workflow.Name,
			"artifact", artifact,
			"artifact-repository", workflow.Status.ArtifactRepositoryRef,
		)
		return "", ""
	}

	return bucket, key
}

// FormatAnnotationPatch generates a raw patch for updating the given annotation.
func FormatAnnotationPatch(annotationKey string, annotationValue string) ([]byte, error) {
	// The annotation key needs to be escaped, since it may contain '/'
	// characters, which already have meaning in the path spec. See
	// https://tools.ietf.org/html/rfc6901#section-3 for more details.
	//
	// Because the characters '~' (%x7E) and '/' (%x2F) have special
	// meanings in JSON Pointer, '~' needs to be encoded as '~0' and '/'
	// needs to be encoded as '~1' when these characters appear in a
	// reference token.
	annotationKey = strings.ReplaceAll(annotationKey, "~", "~0")
	annotationKey = strings.ReplaceAll(annotationKey, "/", "~1")
	path := "/metadata/annotations/" + annotationKey

	//  patch specifies a patch operation for a string.
	payload := []struct {
		Op    string `json:"op"`
		Path  string `json:"path"`
		Value string `json:"value"`
	}{{
		Op:    "replace",
		Path:  path,
		Value: annotationValue,
	}}

	return json.Marshal(payload)
}
