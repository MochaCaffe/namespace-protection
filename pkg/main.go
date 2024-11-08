package main

import (
	"encoding/json"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/http"
	"os"
	"vcluster-gatekeeper/cert"
	"vcluster-gatekeeper/webhook"
)

const (
	// InvalidMessage will be return to the user.
	InvalidMessage = "This namespace contains a Vcluster instance that cannot be destroyed"
	port           = ":8443"
)

// Namespace struct for parsing
type Namespace struct {
	Metadata Metadata `json:"metadata"`
}

// Metadata struct for parsing
type Metadata struct {
	Name        string      `json:"name"`
	Annotations Annotations `json:"annotations"`
}

type Annotations struct {
	Protection string `json:"vClusterProtection"`
}

func (m Metadata) isEmpty() bool {
	return m.Name == ""
}

// Validate handler accepts or rejects based on request contents
func Validate(w http.ResponseWriter, r *http.Request) {
	log.Println("Validating request")
	arReview := &admissionv1.AdmissionReview{}
	if err := json.NewDecoder(r.Body).Decode(&arReview); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err)
		return
	} else if arReview.Request == nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		log.Println(err)
		return
	}

	raw := arReview.Request.OldObject.Raw

	ns := Namespace{
		Metadata: Metadata{
			Annotations: Annotations{},
		},
	}
	if err := json.Unmarshal(raw, &ns); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err)
		log.Println(arReview.Request.OldObject)

		return
	} else if ns.Metadata.isEmpty() {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		log.Println(err)

		return
	}

	arReview.Response = &admissionv1.AdmissionResponse{
		UID:     arReview.Request.UID,
		Allowed: true,
	}

	if ns.Metadata.Annotations.Protection == "true" {
		arReview.Response.Allowed = false
		arReview.Response.Result = &metav1.Status{
			Message: InvalidMessage,
		}
		log.Println(ns.Metadata.Name + ": request refused")

	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&arReview)
	log.Println("Request validated")

}

func main() {
	crt, err := cert.GenCert()
	if err != nil {
		log.Fatal(err)
	}
	crt.SaveCert("/etc/tls")

	_, found := os.LookupEnv("WEBHOOK_SERVICE")
	if !found {
		log.Panic("WEBHOOK_SERVICE environment variable not set")
	}

	_, found = os.LookupEnv("WEBHOOK_NAMESPACE")
	if !found {
		log.Panic("WEBHOOK_NAMESPACE environment variable not set")
	}
	webhook.ApplyValidationConfig(crt)
	http.HandleFunc("/validate", Validate)
	log.Fatal(http.ListenAndServeTLS(port, "/etc/tls/tls.crt", "/etc/tls/tls.key", nil))
}
