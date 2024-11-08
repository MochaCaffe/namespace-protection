package main

import (
	"encoding/json"
	"errors"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"namespace-protection/cert"
	"namespace-protection/webhook"
	"net/http"
	"os"
)

const (
	port = ":8443"
)

// Namespace struct for parsing
type Namespace struct {
	Metadata Metadata `json:"metadata"`
}

// Metadata struct for parsing
type Metadata struct {
	Name        string            `json:"name"`
	Annotations map[string]string `json:"annotations"`
}

func (m Metadata) isEmpty() bool {
	return m.Name == ""
}

// Validate handler accepts or rejects based on request contents
func Validate(w http.ResponseWriter, r *http.Request) {
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
			Annotations: map[string]string{},
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

	rejectionMessage, _ := os.LookupEnv("WEBHOOK_REJECTION_MESSAGE")
	progtectionAnnotation, _ := os.LookupEnv("WEBHOOK_ANNOTATION")
	if ns.Metadata.Annotations[progtectionAnnotation] == "true" {
		arReview.Response.Allowed = false
		arReview.Response.Result = &metav1.Status{
			Message: rejectionMessage,
		}
		log.Println("Delete request refused for namespace " + ns.Metadata.Name)

	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&arReview)
	log.Println("Delete request validated for namespace " + ns.Metadata.Name)

}

func main() {
	serverCert := cert.Certificate{}
	err := serverCert.GenerateSelfSigned()
	if err != nil {
		log.Fatal(err)
	}
	serverCert.SaveCertToPath("/etc/tls")

	err = checkEnvVars([]string{"WEBHOOK_SERVICE", "WEBHOOK_NAMESPACE", "WEBHOOK_REJECTION_MESSAGE", "WEBHOOK_ANNOTATION"})

	if err != nil {
		log.Panic(err)
	}
	webhook.ApplyValidationConfig(&serverCert)
	http.HandleFunc("/validate", Validate)
	log.Fatal(http.ListenAndServeTLS(port, "/etc/tls/tls.crt", "/etc/tls/tls.key", nil))
}

func checkEnvVars(env []string) error {
	for _, envVar := range env {
		if _, found := os.LookupEnv(envVar); !found {
			return errors.New(envVar + "environment variable not set")
		}
	}
	return nil
}
