package escalation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"node-e2e/utils/config"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

var (
	testsEnvironment env.Environment
	saName           string = "node-lister"
	namespace        string = "default"
	crPath           string = "testdata/node-list-cr.yaml"
	prevAcc          *ServiceAccount
	newAcc           *ServiceAccount
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
	if cfg.Namespace() == "" {
		a.WithNamespace(namespace)
	} else {
		a.WithNamespace(cfg.Namespace())
	}

	// Generate a new KubeConfig from passed flags and store the result KC path
	kcPath, err := config.NewKubeConfigFromFlags(a)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	// store the currently used ServiceAccount object
	prevAcc = New(a.GetServiceAccountName(), a.GetServiceAccountNamespace(), a.GetServiceAccountToken())

	// Set the config's KubeConfig
	cfg = cfg.WithKubeconfigFile(kcPath)

	// Finally create the Environment
	testsEnvironment = env.NewWithConfig(cfg)

	// Create a new klient.Client using the previously set KubeConfig
	cfg.NewClient()

	testsEnvironment.Setup(func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		// Create a new ServiceAccount with the provided name and namespace
		s, err := NewServiceAccount(saName, namespace)(ctx, c)
		if err != nil {
			return ctx, err
		}
		newAcc = s

		// Get the absulute path to the ClusterRole to be created
		absCRPath, err := filepath.Abs(crPath)
		if err != nil {
			return ctx, err
		}
		// Assign the ClusterRole to the newly create ServiceAccount using the current, privileged, account
		if err := newAcc.AssignClusterRole(absCRPath)(ctx, c); err != nil {
			return ctx, err
		}

		// Switch to the new ServiceAccount and store the old one
		_, err = SwitchAccount(newAcc)(ctx, c)
		if err != nil {
			return ctx, err
		}

		fmt.Printf("Old token: %s\n", prevAcc.token)
		fmt.Printf("New token: %s\n", c.Client().RESTConfig().BearerToken)

		return ctx, nil
	})

	testsEnvironment.Finish(func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		// Switch to the privileged account and store the unprivileged one
		_, err := SwitchAccount(prevAcc)(ctx, c)
		if err != nil {
			return ctx, err
		}

		fmt.Printf("Old token: %s\n", newAcc.token)
		fmt.Printf("New token: %s\n", c.Client().RESTConfig().BearerToken)

		// Get the absulute path to the ClusterRole to be delete
		absCRPath, err := filepath.Abs(crPath)
		if err != nil {
			return ctx, err
		}
		// Attempt to clean up ServiceAccount, ClusterRole and ClusterRoleBinding
		if err := newAcc.CleanUp(absCRPath)(ctx, c); err != nil {
			return ctx, err
		}
		return ctx, nil
	})

	os.Exit(testsEnvironment.Run(m))
}

func TestPrivilegeEscalation(t *testing.T) {
	feat := features.New("Privilege Escalation").
		WithLabel("type", "Privilege").
		Assess("Test listing all nodes in the cluster using less privileged account", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {

			var nodeList corev1.NodeList
			var secretList corev1.SecretList

			// Try listing all nodes in the cluster using the provided ClusterRole
			if err := c.Client().Resources().List(ctx, &nodeList); err != nil {
				t.Fatal(err)
			}
			for _, node := range nodeList.Items {
				t.Log(node.ObjectMeta.GetName())
			}
			// Try listing other resource
			if err := c.Client().Resources(namespace).List(ctx, &secretList); err != nil {
				t.Logf("could not list Secrets in namespace: %s: %v", c.Namespace(), err)
			} else {
				t.Fatal("account switch was not successful")
			}
			t.Logf("switched account sucessfully")
			return ctx
		}).Feature()

	testsEnvironment.Test(t, feat)
}
