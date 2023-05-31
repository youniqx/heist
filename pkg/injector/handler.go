package injector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/hashicorp/vault/sdk/helper/strutil"
	"github.com/youniqx/heist/pkg/controllers/common"
	"github.com/youniqx/heist/pkg/operator"
	"github.com/youniqx/heist/pkg/vault"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// AnnotationEnableInjection is an annotation used to enable injection of the heist agent into a Pod.
	AnnotationEnableInjection = "heist.youniqx.com/inject-agent"

	// AnnotationAgentStatus is an annotation used by the heist operator to keep track of the injection status in Pods.
	AnnotationAgentStatus = "heist.youniqx.com/agent-status"

	// AnnotationAgentImage is an annotation used to customize the injected agent image.
	AnnotationAgentImage = "heist.youniqx.com/agent-image"

	// AnnotationAgentPreload is an annotation used to customize whether an
	// InitContainer is created to make sure the secret is there before the
	// main container starts.
	AnnotationAgentPreload = "heist.youniqx.com/agent-preload"

	// AnnotationAgentMountPaths is an annotation used to customize paths where secrets can be written.
	AnnotationAgentMountPaths = "heist.youniqx.com/agent-paths"

	// AgentStatusInjected is the value of the AnnotationAgentStatus annotation when the agent has already been injected.
	AgentStatusInjected = "injected"
)

var (
	deserializer = func() runtime.Decoder {
		codecs := serializer.NewCodecFactory(runtime.NewScheme())
		return codecs.UniversalDeserializer()
	}

	kubeSystemNamespaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}
)

type Handler struct {
	Log           logr.Logger
	VaultAPI      vault.API
	K8sClient     client.Client
	Filter        operator.AnnotationFilter
	VaultAddress  string
	AuthMountPath string
	Config        *Config
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Log.Info("Request received", "Method", r.Method, "URL", r.URL)

	if ct := r.Header.Get("Content-Type"); ct != "application/json" {
		msg := fmt.Sprintf("Invalid content-type: %q", ct)
		http.Error(w, msg, http.StatusBadRequest)
		h.Log.Info("warning for request", "Warn", msg, "Code", http.StatusBadRequest)
		return
	}

	var body []byte
	if r.Body != nil {
		var err error
		if body, err = io.ReadAll(r.Body); err != nil {
			msg := fmt.Sprintf("error reading request body: %s", err)
			http.Error(w, msg, http.StatusBadRequest)
			h.Log.Info("error on request", "Error", msg, "Code", http.StatusBadRequest)
			return
		}
	}
	if len(body) == 0 {
		msg := "Empty request body"
		http.Error(w, msg, http.StatusBadRequest)
		h.Log.Info("warning for request", "Warn", msg, "Code", http.StatusBadRequest)
		return
	}

	var admReq v1.AdmissionReview
	var admResp v1.AdmissionReview
	if _, _, err := deserializer().Decode(body, nil, &admReq); err != nil {
		msg := fmt.Sprintf("error decoding admission request: %s", err)
		http.Error(w, msg, http.StatusInternalServerError)
		h.Log.Info("error on request", "Error", msg, "Code", http.StatusInternalServerError)
		return
	}
	admResp.Response = h.Mutate(admReq.Request)
	admResp.APIVersion = "admission.k8s.io/v1"
	admResp.Kind = "AdmissionReview"
	resp, err := json.Marshal(&admResp)
	if err != nil {
		msg := fmt.Sprintf("error marshalling admission response: %s", err)
		http.Error(w, msg, http.StatusInternalServerError)
		h.Log.Info("error on request", "Error", msg, "Code", http.StatusInternalServerError)
		return
	}
	h.Log.Info(string(resp))
	parsed := &struct {
		Response struct {
			Patch string `json:"patch"`
		} `json:"response"`
	}{}

	err = json.Unmarshal(resp, parsed)
	if err != nil {
		msg := fmt.Sprintf("error marshalling admission response: %s", err)
		http.Error(w, msg, http.StatusInternalServerError)
		h.Log.Info("error on request", "Error", msg, "Code", http.StatusInternalServerError)
		return
	}

	h.Log.Info("finished handling pod webhook")
	if _, err := w.Write(resp); err != nil {
		h.Log.Info("error writing response", "Error", err)
	}
}

// Mutate takes an admission request and performs mutation if necessary,
// returning the final API response.
func (h *Handler) Mutate(req *v1.AdmissionRequest) *v1.AdmissionResponse {
	// Decode the pod from the request
	var pod corev1.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		h.Log.Info("could not unmarshal request to pod: %s", err)
		return &v1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	// Build the basic response
	resp := &v1.AdmissionResponse{
		Allowed: true,
		UID:     req.UID,
	}

	inject, err := h.ShouldInject(&pod)
	if err != nil && !strings.Contains(err.Error(), "no inject annotation found") {
		err := fmt.Errorf("error checking if should inject agent: %w", err)
		return admissionError(err)
	} else if !inject {
		h.Log.Info("not injecting vault sidecar agents")
		return resp
	}

	if strutil.StrListContains(kubeSystemNamespaces, req.Namespace) {
		err := fmt.Errorf("error with request namespace: cannot inject into system namespaces: %s", req.Namespace)
		return admissionError(err)
	}

	injector, err := h.NewInjector(&pod)
	if err != nil {
		err := fmt.Errorf("error creating new agent sidecar: %w", err)
		return admissionError(err)
	}

	patches, err := injector.Patch()
	if err != nil {
		err := fmt.Errorf("error creating patch for agent: %w", err)
		return admissionError(err)
	}

	patchData, err := json.Marshal(patches)
	if err != nil {
		err := fmt.Errorf("error encoding patches for agent: %w", err)
		return admissionError(err)
	}

	resp.Patch = patchData
	patchType := v1.PatchTypeJSONPatch
	resp.PatchType = &patchType
	return resp
}

func admissionError(err error) *v1.AdmissionResponse {
	return &v1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

func (h *Handler) ShouldInject(pod *corev1.Pod) (bool, error) {
	if !h.Filter.Matches(pod) {
		return false, nil
	}

	raw, ok := common.GetAnnotationValue(pod, AnnotationEnableInjection)
	if !ok {
		return false, nil
	}

	shouldInject, err := strconv.ParseBool(raw)
	if err != nil {
		h.Log.Info("couldn't parse agent injection annotation", "error", err)
		return false, err
	}

	if !shouldInject {
		return false, nil
	}

	switch status, _ := common.GetAnnotationValue(pod, AnnotationAgentStatus); status {
	case AgentStatusInjected:
		return false, nil
	default:
		return true, nil
	}
}
