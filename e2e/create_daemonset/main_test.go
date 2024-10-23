package createdaemonset_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"node-e2e/utils/config"
	"node-e2e/utils/template"

	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

var (
	testsEnvironment env.Environment
	nsPrefix         string = "node-e2e"
	namespace        string = envconf.RandomName(nsPrefix, 13)
)

func TestMain(m *testing.M) {
	// Build Environment configuration from provided flags to allow tests filtering and other capabilities
	cfg, err := envconf.NewFromFlags()
	if err != nil {
		fmt.Printf("failed to build environment configuration from flags: %s", err)
		os.Exit(1)
	}

	// Create a new AuthenticationAttr object
	a := config.New()
	// Set namespace for later KubeConfig generation
	a.WithNamespace(namespace)

	// Generate a new KubeConfig from passed flags and store the result KC path
	kcPath, err := config.NewKubeConfigFromFlags(a)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	// Set the config's KubeConfig
	cfg = cfg.WithKubeconfigFile(kcPath)

	// Finally create the Environment
	testsEnvironment = env.NewWithConfig(cfg)

	// Create a new klient.Client using the previously set KubeConfig
	cfg.NewClient()

	os.Exit(testsEnvironment.Run(m))
}

func TestResourceCreationFromFiles(t *testing.T) {
	feat := features.New("Decoder").
		WithLabel("type", "VM").
		Assess("Test creating resources using decoder", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// Read the template file with placeholders
			templateData, err := os.ReadFile("testdata/daemon-manager-crb.yaml")
			if err != nil {
				t.Fatal(err)
			}
			// Fill the blanks
			templatedCRB, err := template.Template(map[string]interface{}{"namespace": namespace}, templateData)
			if err != nil {
				t.Fatal(err)
			}

			crb := rbacv1.ClusterRoleBinding{}
			// Create the different resources
			if err := decoder.Decode(templatedCRB, &crb); err != nil {
				t.Fatal(err)
			}

			if err := c.Client().Resources().Create(ctx, &crb); err != nil {
				t.Fatal(err)
			}

			return ctx
		}).Feature()

	testsEnvironment.Test(t, feat)
}
