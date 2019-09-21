package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	"k8s.io/klog/klogr"
	"k8s.io/test-infra-setup/webhook/pkg/kustomize"
	"net/http"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/yaml"
	"text/template"
)

var (
	scheme = runtime.NewScheme()
	log    = ctrl.Log.WithName("setup")

	parsedTemplates []*template.Template
)

func main() {
	// Setting up logger flags
	klog.InitFlags(nil)

	var (
		metricsAddr string
		certDir     string
		patchesGlob string
	)
	flag.StringVar(
		&metricsAddr,
		"metrics-addr",
		":8080",
		"The address the metric endpoint binds to.",
	)
	flag.StringVar(
		&certDir,
		"cert-dir",
		"/tmp/k8s-webhook-server/serving-certs",
		"The folder were tls.crt and tls.key are stored.",
	)
	flag.StringVar(
		&patchesGlob,
		"patches-glob",
		"/tmp/patches/*",
		"The glob pattern to parse the patches from.",
	)
	log.Info("Parsing flags")
	flag.Parse()

	ctrl.SetLogger(klogr.New())

	log.Info("Parsing templates")
	parsedTemplates = template.Must(template.ParseGlob(patchesGlob)).Templates()

	log.Info("Setting up manager")
	cfg, err := config.GetConfigWithContext(os.Getenv("KUBECONTEXT"))
	if err != nil {
		log.Error(err, "unable to get kubeconfig")
		os.Exit(1)
	}
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	log.Info("Setting up webhook server")
	webhookServer := mgr.GetWebhookServer()
	webhookServer.Register("/mutate", &admission.Webhook{
		Handler: &mutatingHandler{},
	})
	webhookServer.CertDir = certDir
	webhookServer.Port = 8443

	log.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")
		os.Exit(1)
	}
}

type mutatingHandler struct {
	decoder *admission.Decoder
}

var _ admission.DecoderInjector = &mutatingHandler{}

// InjectDecoder injects the decoder into a mutatingHandler.
func (h *mutatingHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

// Handle handles admission requests.
func (h *mutatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {

	// Decode the object in the request and set TypeMeta for kustomize
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
	}
	err := h.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Render patches with go template with v1.Pod as data
	var patches []string
	for _, t := range parsedTemplates {
		var p bytes.Buffer
		err = t.Execute(&p, pod)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		patches = append(patches, p.String())
	}

	// Marshal Pod to bytes
	objBytes, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Execute kustomize build
	newPod, err := kustomize.Build([]string{string(objBytes)}, patches, []kustomize.PatchJSON6902{})
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Convert patched Pod from YAML to Json
	newPodJson, err := yaml.YAMLToJSON([]byte(newPod))
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Patch current Pod to newPod
	return admission.PatchResponseFromRaw(req.Object.Raw, newPodJson)
}
