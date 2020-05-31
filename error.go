package webhook

import (
	"k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
)

func errorToAdmissionResponse(err error) *v1.AdmissionResponse {
	return &v1.AdmissionResponse{Result: &metav1.Status{Message: err.Error()}}
}

func handleError(message string) {
	log.Print(message)
}
