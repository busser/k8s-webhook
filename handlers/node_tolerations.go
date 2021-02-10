package handlers

import (
	"encoding/json"
	"fmt"
	"log"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TolerationKey is the injected toleration's key.
const TolerationKey = "padok.fr/namespace"

// AddTolerations responds to an AdmissionRequest for a Pod with a patch that
// adds a toleration based on the pod's namespace.
func AddTolerations(request admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	gvk := metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}
	if request.Kind != gvk {
		err := fmt.Errorf("expected %s, got %s", gvk, request.Kind)
		return admissionResponseError(err)
	}

	var pod corev1.Pod
	err := json.Unmarshal(request.Object.Raw, &pod)
	if err != nil {
		return admissionResponseError(err)
	}

	log.Printf("Processing Pod with name %q in namespace %q", request.Name, request.Namespace)

	var patch JSONPatch

	// If the Pod has no tolerations, set an initial value: an empty slice of
	// tolerations.
	if len(pod.Spec.Tolerations) == 0 {
		patch = append(patch, JSONPatchOperation{
			Op:    "add",
			Path:  "/spec/tolerations",
			Value: []corev1.Toleration{},
		})
	}

	// This is the toleration we want to add.
	toleration := corev1.Toleration{
		Key:    TolerationKey,
		Value:  request.Namespace,
		Effect: corev1.TaintEffectNoSchedule,
	}

	var hasToleration bool
	for _, t := range pod.Spec.Tolerations {
		if t == toleration {
			hasToleration = true
		}
	}

	if !hasToleration {
		patch = append(patch, JSONPatchOperation{
			Op:    "add",
			Path:  fmt.Sprintf("/spec/tolerations/%d", len(pod.Spec.Tolerations)),
			Value: toleration,
		})
	}

	jsonPatch, err := json.Marshal(patch)
	if err != nil {
		return admissionResponseError(err)
	}

	var response admissionv1.AdmissionResponse
	response.UID = request.UID
	response.Allowed = true
	response.Patch = jsonPatch
	patchType := admissionv1.PatchTypeJSONPatch
	response.PatchType = &patchType

	return &response
}

func admissionResponseError(err error) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}
