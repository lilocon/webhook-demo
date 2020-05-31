package webhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/http"
)

func createPatch(pod *corev1.Pod) ([]byte, error) {
	var operations []rfc6902PatchOperation

	for i, container := range pod.Spec.Containers {
		cpuRequest := container.Resources.Requests.Cpu().Format

		fmt.Println(cpuRequest)

		path := fmt.Sprintf("/spec/containers/%d", i)
		operations = append(operations, replaceCpuRequest(container, path))
	}

	return json.Marshal(operations)
}

func inject(ar *v1.AdmissionReview) *v1.AdmissionResponse {
	req := ar.Request
	var pod corev1.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		handleError(fmt.Sprintf("Could not unmarshal raw object: %v %s", err,
			string(req.Object.Raw)))
		return errorToAdmissionResponse(err)
	}

	// Deal with potential empty fields, e.g., when the pod is created by a deployment
	podName := potentialPodName(&pod.ObjectMeta)
	if pod.ObjectMeta.Namespace == "" {
		pod.ObjectMeta.Namespace = req.Namespace
	}

	log.Printf("AdmissionReview for Kind=%v Namespace=%v Name=%v (%v) UID=%v Rfc6902PatchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, podName, req.UID, req.Operation, req.UserInfo)
	log.Printf("Object: %v", string(req.Object.Raw))
	log.Printf("OldObject: %v", string(req.OldObject.Raw))

	if !injectRequired(&pod.ObjectMeta) {
		log.Printf("Skipping %s/%s due to policy check", pod.ObjectMeta.Namespace, podName)
		return &v1.AdmissionResponse{
			Allowed: true,
		}
	}

	deployMeta := pod.ObjectMeta.DeepCopy()
	deployMeta.Namespace = req.Namespace

	patchBytes, err := createPatch(&pod)
	if err != nil {
		handleError(fmt.Sprintf("AdmissionResponse: err=%v\n", err))
		return errorToAdmissionResponse(err)
	}

	log.Printf("AdmissionResponse: patch=%v\n", string(patchBytes))

	reviewResponse := v1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1.PatchType {
			pt := v1.PatchTypeJSONPatch
			return &pt
		}(),
	}
	return &reviewResponse
}

func Serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if len(body) == 0 {
		handleError("no body found")
		http.Error(w, "no body found", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		handleError(fmt.Sprintf("contentType=%s, expect application/json", contentType))
		http.Error(w, "invalid Content-Type, want `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	requestedAdmissionReview := v1.AdmissionReview{}
	responseAdmissionReview := v1.AdmissionReview{}

	if _, _, err := deserializer.Decode(body, nil, &requestedAdmissionReview); err != nil {
		handleError(fmt.Sprintf("Could not decode body: %v", err))
		responseAdmissionReview.Response = errorToAdmissionResponse(err)
	} else {
		responseAdmissionReview.Response = inject(&requestedAdmissionReview)
	}

	responseAdmissionReview.APIVersion = "admission.k8s.io/v1"
	responseAdmissionReview.Kind = "AdmissionReview"
	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID

	resp, err := json.Marshal(responseAdmissionReview)

	if err != nil {
		log.Printf("Could not encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}

	log.Printf("Sending Response: %v", responseAdmissionReview.Response)

	if _, err := w.Write(resp); err != nil {
		log.Printf("Could not write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}

func potentialPodName(metadata *metav1.ObjectMeta) string {
	if metadata.Name != "" {
		return metadata.Name
	}
	if metadata.GenerateName != "" {
		return metadata.GenerateName + "***** (actual name not yet known)"
	}
	return ""
}

func injectRequired(metadata *metav1.ObjectMeta) bool {

	return true
}
