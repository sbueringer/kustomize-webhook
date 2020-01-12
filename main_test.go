package main

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"text/template"

	. "github.com/onsi/ginkgo"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
)

var _ = Describe("Mutating webhook", func() {
	It("TODO", func() {
		By("invoking the webhook")


		parsedTemplates = []*template.Template{template.Must(template.New("test").Parse(`
apiVersion: v1
kind: Pod
metadata:
  name: {{.Name}}
  namespace: test-pods
spec:
  {{if index .Annotations "pvc.daimler.com/size" }}
  volumes:
  - name: build
	persistentVolumeClaim:
	  claimName: {{.Name}}
  {{ end }}
		`))}

		resp := getMutatingHandler().Handle(context.Background(), admission.Request{
			AdmissionRequest: admissionv1beta1.AdmissionRequest{
				UID: "foobar",
				Object: runtime.RawExtension{
					Raw: mustMarshal(json.Marshal(&v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test-pods",
							Labels: map[string]string{
								"inject": "tproxy",
							},
							Annotations: map[string]string{
								"pvc.daimler.com/size": "15Gi",
							},
						},
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "test",
									Image: "bazel:v1",
								},
							},
						},
					})),
				},
			},
		})

		fmt.Printf("%+v", resp.Patches)
		//By("checking that the response share's the request's UID")
		//Expect(resp.UID).To(Equal(machinerytypes.UID("foobar")))
	})
})

func mustMarshal(bytes []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return bytes
}

func getMutatingHandler() *mutatingHandler {
	mwh := &mutatingHandler{}
	decoder, err := admission.NewDecoder(runtime.NewScheme())
	if err != nil {
		panic(err)
	}
	err = mwh.InjectDecoder(decoder)
	if err != nil {
		panic(err)
	}
	return mwh
}
