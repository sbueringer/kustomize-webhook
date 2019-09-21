package kustomize

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sigs.k8s.io/kustomize/v3/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/v3/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/v3/k8sdeps/validator"
	"sigs.k8s.io/kustomize/v3/pkg/commands/build"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
)

type PatchJSON6902 struct {
	// these fields specify the patch target resource
	Group   string
	Version string
	Kind    string
	// Name and Namespace are optional
	// NOTE: technically name is required now, but we default it elsewhere
	// Third party users of this type / library would need to set it.
	Name      string
	Namespace string
	// Patch should contain the contents of the json patch as a string
	Patch string
}

// Build takes a set of resource blobs (yaml), patches (strategic merge patch)
// https://github.com/kubernetes/community/blob/master/contributors/devel/strategic-merge-patch.md
// and returns the `kustomize build` result as a yaml blob
// It does this in-memory using the build cobra command
// Copied from https://github.com/kubernetes-sigs/kind/blob/master/pkg/internal/util/kustomize/kustomize.go
func Build(resources, patches []string, patchesJSON6902 []PatchJSON6902) (string, error) {
	// write the resources and patches to an in memory fs with a generated
	// kustomization.yaml
	memFS := fs.MakeFakeFS()
	var kustomization bytes.Buffer
	fakeDir := "/"

	// NOTE: we always write this header as you cannot build without any resources
	kustomization.WriteString("resources:\n")
	for i, res := range resources {
		// this cannot error per docs
		name := fmt.Sprintf("resource-%d.yaml", i)
		_ = memFS.WriteFile(filepath.Join(fakeDir, name), []byte(res))
		fmt.Fprintf(&kustomization, " - %s\n", name)
	}

	if len(patches) > 0 {
		kustomization.WriteString("patches:\n")
	}
	for i, patch := range patches {
		// this cannot error per docs
		name := fmt.Sprintf("patch-%d.yaml", i)
		_ = memFS.WriteFile(filepath.Join(fakeDir, name), []byte(patch))
		fmt.Fprintf(&kustomization, " - %s\n", name)
	}

	if len(patchesJSON6902) > 0 {
		kustomization.WriteString("patchesJson6902:\n")
	}
	for i, patch := range patchesJSON6902 {
		// this cannot error per docs
		name := fmt.Sprintf("patch-json6902-%d.yaml", i)
		_ = memFS.WriteFile(filepath.Join(fakeDir, name), []byte(patch.Patch))
		fmt.Fprintf(&kustomization, " - path: %s\n", name)
		fmt.Fprintf(&kustomization, "   target:\n")
		fmt.Fprintf(&kustomization, "     group: %s\n", patch.Group)
		fmt.Fprintf(&kustomization, "     version: %s\n", patch.Version)
		fmt.Fprintf(&kustomization, "     kind: %s\n", patch.Kind)
		if patch.Name != "" {
			fmt.Fprintf(&kustomization, "     name: %s\n", patch.Name)
		}
		if patch.Namespace != "" {
			fmt.Fprintf(&kustomization, "     namespace: %s\n", patch.Namespace)
		}
	}

	if err := memFS.WriteFile(
		filepath.Join(fakeDir, "kustomization.yaml"), kustomization.Bytes(),
	); err != nil {
		return "", err
	}

	// now we can build the kustomization
	var out bytes.Buffer
	uf := kunstruct.NewKunstructuredFactoryImpl()
	pf := transformer.NewFactoryImpl()
	rf := resmap.NewFactory(resource.NewFactory(uf), pf)
	v := validator.NewKustValidator()
	cmd := build.NewCmdBuild(&out, memFS, v, rf, pf)
	cmd.SetArgs([]string{"--", fakeDir})
	// we want to silence usage, error output, and any future output from cobra
	// we will get error output as a golang error from execute
	cmd.SetOutput(os.Stdout)
	_, err := cmd.ExecuteC()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}