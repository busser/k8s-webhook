package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/busser/k8s-webhook/handlers"
	admissionv1 "k8s.io/api/admission/v1"
)

func main() {
	http.HandleFunc("/inject-tolerations", handlerFrom(handlers.AddTolerations))

	log.Println("Listening on port 8443...")
	err := http.ListenAndServeTLS(
		":8443",
		"/tmp/toleration-injector/serving-certs/tls.crt",
		"/tmp/toleration-injector/serving-certs/tls.key",
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
}

type admitterFunc func(admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse

func handlerFrom(admitter admitterFunc) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		// verify the content type is accurate
		contentType := req.Header.Get("Content-Type")
		if contentType != "application/json" {
			http.Error(resp, "expected Content-Type application/json", http.StatusUnsupportedMediaType)
			return
		}

		// decode the admission review from the HTTP request
		var reviewRequest admissionv1.AdmissionReview
		if err := json.NewDecoder(req.Body).Decode(&reviewRequest); err != nil {
			http.Error(resp, err.Error(), http.StatusBadRequest)
			return
		}

		// provide a response to the admission review's request
		var reviewResponse admissionv1.AdmissionReview
		gvk := admissionv1.SchemeGroupVersion.WithKind("AdmissionReview")
		reviewResponse.SetGroupVersionKind(gvk) // ! Shouldn't this line be after the next one? Is it even necessary?

		reviewResponse.Response = admitter(*reviewRequest.Request)

		// encode the admission review into the HTTP response
		resp.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(resp).Encode(reviewResponse); err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
		}
	}
}
