package handlers

import (
	"encoding/json"
	"fmt"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TolerationKey is the injected toleration's key.
const TolerationKey = "padok.fr/namespace"

// AddTolerations responds to an AdmissionRequest for a Pod with a patch that
// adds a toleration based on the pod's namespace.
func AddTolerations(request admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Make sure that the request's Kind is for a Pod resource.
	gvk := metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}
	if request.Kind != gvk {
		err := fmt.Errorf("expected %s, got %s", gvk, request.Kind)
		return admissionResponseError(err)
	}

	// Decode the Pod from the request.
	var pod corev1.Pod
	err := json.Unmarshal(request.Object.Raw, &pod)
	if err != nil {
		return admissionResponseError(err)
	}

	// Prepare a JSON patch.
	var patch JSONPatch

	// If the Pod has no tolerations, set an initial value: an empty slice of
	// tolerations.
	if pod.Spec.Tolerations == nil {
		patch.Append(JSONPatchOperation{
			Op:    "add",
			Path:  "/spec/tolerations",
			Value: []corev1.Toleration{},
		})
	}

	// Define the toleration to add.
	toleration := corev1.Toleration{
		Key:    TolerationKey,
		Value:  request.Namespace,
		Effect: corev1.TaintEffectNoSchedule,
	}

	// Check if the pod already has the toleration.
	var hasToleration bool
	for _, t := range pod.Spec.Tolerations {
		if t == toleration {
			hasToleration = true
		}
	}

	// Add the toleration if it is missing.
	if !hasToleration {
		patch.Append(JSONPatchOperation{
			Op:    "add",
			Path:  "/spec/tolerations/-",
			Value: toleration,
		})
	}

	// Encode the patch as JSON.
	encodedPatch, err := json.Marshal(patch)
	if err != nil {
		return admissionResponseError(err)
	}

	// Prepare a response.
	var response admissionv1.AdmissionResponse
	response.UID = request.UID
	response.Allowed = true

	// Include the patch in the response.
	response.Patch = encodedPatch
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
