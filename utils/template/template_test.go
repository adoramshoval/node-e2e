package template

import (
	"context"
	"fmt"
	"os"
	"testing"

	"node-e2e/utils/escalation"
	"node-e2e/utils/tests"

	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

var (
	testsEnvironment env.Environment
	nsPrefix         string = "node-e2e"
	namespace        string = envconf.RandomName(nsPrefix, 13)
	privAcc          *escalation.ServiceAccount
)

func TestMain(m *testing.M) {
	e, a, err := tests.StartWithServiceAccountFlags(namespace)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	testsEnvironment = e
	privAcc = a

	os.Exit(testsEnvironment.Run(m))
}

func TestResourceCreationFromFiles(t *testing.T) {
	feat := features.New("Template").
		WithLabel("type", "template").
		Assess("Test templating YAML files", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// Read the template file with placeholders
			templateData, err := os.ReadFile("testdata/daemon-manager-crb.yaml")
			if err != nil {
				t.Fatal(err)
			}
			// Fill the blanks
			templatedCRB, err := Template(map[string]interface{}{"namespace": namespace}, templateData)
			if err != nil {
				t.Fatal(err)
			}

			crb := rbacv1.ClusterRoleBinding{}
			// Create the different resources
			if err := decoder.Decode(templatedCRB, &crb); err != nil {
				t.Fatal(err)
			}

			if err := c.Client().Resources().Create(ctx, &crb); !apierrors.IsAlreadyExists(err) {
				t.Fatal(err)
			}

			if crb.Subjects[0].Namespace != namespace {
				t.Fatalf("subject namespace not as expected, template failed: expected %s, got %s", namespace, crb.Subjects[0].Namespace)
			}

			t.Log("template succeeded")

			return ctx
		}).Feature()

	testsEnvironment.Test(t, feat)
}
