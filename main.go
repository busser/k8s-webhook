package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/busser/k8s-webhook/handlers"
	admissionv1 "k8s.io/api/admission/v1"
)

func main() {
	http.HandleFunc("/inject-tolerations", httpHandlerFrom(handlers.AddTolerations))

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

// AdmissionHandlerFunc responds to a Kubernetes admission request.
type AdmissionHandlerFunc func(admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse

func httpHandlerFrom(handler AdmissionHandlerFunc) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		// Verify the content type is accurate.
		contentType := req.Header.Get("Content-Type")
		if contentType != "application/json" {
			http.Error(resp, "expected Content-Type application/json", http.StatusUnsupportedMediaType)
			return
		}

		// Decode the admission review from the HTTP request.
		var review admissionv1.AdmissionReview
		if err := json.NewDecoder(req.Body).Decode(&review); err != nil {
			http.Error(resp, err.Error(), http.StatusBadRequest)
			return
		}

		// Call the admission handler.
		review.Response = handler(*review.Request)

		// Encode the admission review into the HTTP response
		resp.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(resp).Encode(review); err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
		}
	}
}
