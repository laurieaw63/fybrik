// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/helm"
	"fybrik.io/fybrik/pkg/logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"
)

func readBlueprint(f string) (*app.Blueprint, error) {
	blueprintYAML, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	blueprint := &app.Blueprint{}
	err = yaml.Unmarshal(blueprintYAML, blueprint)
	if err != nil {
		return nil, err
	}
	return blueprint, nil
}

func TestBlueprintReconcile(t *testing.T) {
	blueprintNamespace := utils.GetSystemNamespace()
	fmt.Printf("Blueprint controller unit test: Using blueprint namespace: %s\n", blueprintNamespace)

	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	blueprint, err := readBlueprint("../../testdata/blueprint.yaml")
	g.Expect(err).To(gomega.BeNil(), "Cannot read blueprint file for test")
	blueprint.Namespace = blueprintNamespace
	blueprint.Spec.ModulesNamespace = utils.GetDefaultModulesNamespace()

	// Objects to track in the fake client.
	objs := []runtime.Object{
		blueprint,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	r := &BlueprintReconciler{
		Client: cl,
		Name:   "BlueprintTestController",
		Log:    logging.LogInit(logging.CONTROLLER, "test-blueprint-controller"),
		Scheme: s,
		Helmer: helm.NewEmptyFake(),
	}
	ns := client.ObjectKeyFromObject(blueprint)

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: ns,
	}

	res, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	// Check the result of reconciliation to make sure it has the desired state.
	g.Expect(res.Requeue).To(gomega.BeFalse(), "reconcile did not requeue request as expected")
	g.Expect(cl.Get(context.TODO(), ns, blueprint)).To(gomega.BeNil(), "could not fetch the blueprint")
	g.Expect(blueprint.Status.Releases).To(gomega.HaveLen(2))
	g.Expect(blueprint.Status.Releases).Should(gomega.HaveKeyWithValue("notebook-default-notebook-copy-batch", blueprint.Status.ObservedGeneration))
	g.Expect(blueprint.Status.Releases).Should(gomega.HaveKeyWithValue("notebook-default-notebook-read-module", blueprint.Status.ObservedGeneration))
}

// This test checks that a short release name is not truncated
func TestShortReleaseName(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	modules := map[string]app.BlueprintModule{"dataFlowInstance1": {
		Name:      "dataFlow",
		Chart:     app.ChartSpec{Name: "thechart"},
		Arguments: app.ModuleArguments{},
	}}

	blueprint := app.Blueprint{
		ObjectMeta: metav1.ObjectMeta{
			Name: "appns-app-mybp",
			Labels: map[string]string{
				app.ApplicationNameLabel:      "my-app",
				app.ApplicationNamespaceLabel: "default",
			},
		},
		Spec: app.BlueprintSpec{
			Cluster: "cluster1",
			Modules: modules,
		},
	}
	relName := utils.GetReleaseName(blueprint.Labels[app.ApplicationNameLabel], blueprint.Labels[app.ApplicationNamespaceLabel], "dataFlowInstance1")
	g.Expect(relName).To(gomega.Equal("my-app-default-dataFlowInstance1"))
}

// This test checks that a long release name is shortened
func TestLongReleaseName(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	blueprint := app.Blueprint{
		ObjectMeta: metav1.ObjectMeta{
			Name: "appnsisalreadylong-appnameisevenlonger-myblueprintnameisreallytakingitoverthetopkubernetescantevendealwithit",
			Labels: map[string]string{
				app.ApplicationNameLabel:      "my-app",
				app.ApplicationNamespaceLabel: "default",
			},
		},
		Spec: app.BlueprintSpec{
			Cluster: "cluster1",
			Modules: map[string]app.BlueprintModule{"ohandnottoforgettheflowstepnamethatincludesthetemplatenameandotherstuff": {Name: "longname", Chart: app.ChartSpec{Name: "start-image"}}},
		},
	}

	relName := utils.GetReleaseName(blueprint.Labels[app.ApplicationNameLabel], blueprint.Labels[app.ApplicationNamespaceLabel], "ohandnottoforgettheflowstepnamethatincludesthetemplatenameandotherstuff")
	g.Expect(relName).To(gomega.Equal("my-app-default-ohandnottoforgettheflowstepnamet-99207"))
	g.Expect(relName).To(gomega.HaveLen(53))

	// Make sure that calling the same method again results in the same result
	relName2 := utils.GetReleaseName(blueprint.Labels[app.ApplicationNameLabel], blueprint.Labels[app.ApplicationNamespaceLabel], "ohandnottoforgettheflowstepnamethatincludesthetemplatenameandotherstuff")
	g.Expect(relName2).To(gomega.Equal("my-app-default-ohandnottoforgettheflowstepnamet-99207"))
	g.Expect(relName2).To(gomega.HaveLen(53))
}
