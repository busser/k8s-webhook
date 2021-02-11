package handlers_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/busser/k8s-webhook/handlers"
	"github.com/google/go-cmp/cmp"
	admissionv1 "k8s.io/api/admission/v1"
)

func TestAddTolerations(t *testing.T) {
	testCases := []struct {
		requestFile, responseFile string
	}{
		{
			requestFile:  "testdata/has-toleration/request.json",
			responseFile: "testdata/has-toleration/response.json",
		},
		{
			requestFile:  "testdata/missing-toleration/request.json",
			responseFile: "testdata/missing-toleration/response.json",
		},
		{
			requestFile:  "testdata/no-tolerations/request.json",
			responseFile: "testdata/no-tolerations/response.json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.requestFile, func(t *testing.T) {
			var request admissionv1.AdmissionRequest
			if err := fromJSONFile(tc.requestFile, &request); err != nil {
				t.Fatal(err)
			}

			var expectedResponse admissionv1.AdmissionResponse
			if err := fromJSONFile(tc.responseFile, &expectedResponse); err != nil {
				t.Fatal(err)
			}

			actualResponse := handlers.AddTolerations(request)
			if diff := cmp.Diff(&expectedResponse, actualResponse); diff != "" {
				t.Errorf("AddTolerations() mismatch (-expected +actual):\n%s", diff)
			}
		})
	}
}

func fromJSONFile(file string, obj interface{}) error {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, obj); err != nil {
		return err
	}

	return nil
}
