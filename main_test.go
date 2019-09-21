package main

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	. "github.com/onsi/ginkgo"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
)

var _ = Describe("Mutating webhook", func() {
	It("TODO", func() {
		By("invoking the webhook")

		resp := getMutatingHandler().Handle(context.Background(), admission.Request{
			AdmissionRequest: admissionv1beta1.AdmissionRequest{
				UID: "foobar",
				Object: runtime.RawExtension{
					Raw: mustMarshal(json.Marshal(&v1.Pod{
						ObjectMeta: metav1.ObjectMeta {
							Name: "test",
							Namespace: "test-pods",
							Labels: map[string]string{
								"inject": "tproxy",
							},
						},
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name: "test",
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
	err= mwh.InjectDecoder(decoder)
	if err != nil {
		panic(err)
	}
	return mwh
}
