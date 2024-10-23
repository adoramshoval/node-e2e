### Test Skeleton for OCPBM e2e Testing

#### Overview

The Kubernetes e2e-framework is designed to simplify writing end-to-end tests for Kubernetes environments. It enables the direct interaction with Kubernetes clusters via the Kubernetes API, allowing you to create resources, validate assertions, and manage teardown processes for resources that the test depends on.

This document outlines a suggested structure for writing such tests, ensuring consistency and clarity while maintaining the flexibility to adapt tests to various scenarios.

### Generic Test Structure

The following example demonstrates how to structure a privilege escalation test using the e2e-framework. It ensures that tests can run with a reduced privilege account, assigning only the necessary permissions to the ServiceAccount that the test uses.

```go
package privilege_escalation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"node-e2e/utils/config"
	"node-e2e/utils/escalation"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

var (
	testsEnvironment env.Environment
	saName           = "node-lister"
	namespace        = "default"
	crPath           = "testdata/node-list-cr.yaml"
	prevAcc          *ServiceAccount
	newAcc           *ServiceAccount
)

func TestMain(m *testing.M) {
	// Set up test environment using e2e-framework's envconf.Config
	cfg, err := envconf.NewFromFlags()
	if err != nil {
		fmt.Printf("failed to build environment configuration from flags: %s", err)
		os.Exit(1)
	}

	// Create a new AuthenticationAttr object for the ServiceAccount
	a := config.New()
	// Set namespace for later KubeConfig generation - a flag will overwrite the provided namespace
	if cfg.Namespace() == "" {
		a.WithNamespace(namespace)
	} else {
		a.WithNamespace(cfg.Namespace())
	}

	// Generate a KubeConfig based on provided flags
	kcPath, err := config.NewKubeConfigFromFlags(a)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	// Store the ServiceAccount object
	prevAcc = escalation.New(a.GetServiceAccountName(), a.GetServiceAccountNamespace(), a.GetServiceAccountToken())

	// Apply KubeConfig to the test environment
	cfg = cfg.WithKubeconfigFile(kcPath)
	testsEnvironment = env.NewWithConfig(cfg)
	cfg.NewClient()

	// Setup and cleanup for tests
	testsEnvironment.Setup(func(ctx context.Context, c *envconf.Config) (context.Context, error) {
			setupTestEnvironment(ctx, c)
		})
	testsEnvironment.Finish(func(ctx context.Context, c *envconf.Config) (context.Context, error) {
			teardownTestEnvironment(ctx, c)
		})

	// Exit and start test execution
	os.Exit(testsEnvironment.Run(m))
}

func setupTestEnvironment(ctx context.Context, c *envconf.Config) (context.Context, error) {
	// Create a ServiceAccount and assign a ClusterRole
	s, err := escalation.NewServiceAccount(saName, namespace)(ctx, c)
	if err != nil {
		return ctx, err
	}
	newAcc = s

	absCRPath, err := filepath.Abs(crPath)
	if err != nil {
		return ctx, err
	}
	// Create the passed ClusterRole and assign the given permissions to the new ServiceAccount
	err = newAcc.AssignClusterRole(absCRPath)(ctx, c)
	if err != nil {
		return ctx, err
	}

	// Switch to the new ServiceAccount
	_, err = escalation.SwitchAccount(newAcc)(ctx, c)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func teardownTestEnvironment(ctx context.Context, c *envconf.Config) (context.Context, error) {
	// Switch back to the original ServiceAccount and clean up resources
	_, err := escalation.SwitchAccount(prevAcc)(ctx, c)
	if err != nil {
		return ctx, err
	}

	absCRPath, err := filepath.Abs(crPath)
	if err != nil {
		return ctx, err
	}
	// Clean up the ClusterRole, ClusterRoleBinding and ServiceAccount
	err = newAcc.CleanUp(absCRPath)(ctx, c)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func TestPrivilegeEscalation(t *testing.T) {
	feat := features.New("Privilege Escalation").
		WithLabel("type", "Privilege").
		Assess("Test listing all nodes in the cluster using less privileged account", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {

			var nodeList corev1.NodeList
			if err := c.Client().Resources().List(ctx, &nodeList); err != nil {
				t.Fatal(err)
			}
			for _, node := range nodeList.Items {
				t.Log(node.ObjectMeta.GetName())
			}
			// Try listing other resource - should not be able to perform
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
```

#### Example Test Output
```
=== RUN   TestPrivilegeEscalation
=== RUN   TestPrivilegeEscalation/Privilege_Escalation
=== RUN   TestPrivilegeEscalation/Privilege_Escalation/Test_listing_all_nodes_in_the_cluster_using_less_privileged_account
    escalation_test.go:127: control-a-1.medone-1.med.one
    escalation_test.go:127: control-b-1.medone-1.med.one
    escalation_test.go:127: control-c-1.medone-1.med.one
    escalation_test.go:127: phmowrk-166018-14.medone-1.med.one
    escalation_test.go:127: phmowrk-166018-2-4.medone-1.med.one
    escalation_test.go:127: phmowrk-166019-15.medone-1.med.one
    escalation_test.go:131: could not list Secrets in namespace: : secrets is forbidden: User "system:serviceaccount:default:node-lister" cannot list resource "secrets" in API group "" in the namespace "default"
    escalation_test.go:133: switched account sucessfully
--- PASS: TestPrivilegeEscalation (0.03s)
    --- PASS: TestPrivilegeEscalation/Privilege_Escalation (0.03s)
        --- PASS: TestPrivilegeEscalation/Privilege_Escalation/Test_listing_all_nodes_in_the_cluster_using_less_privileged_account (0.03s)
```

#### Detailed Steps:

#### Steps and Functions in the Test

1. **TestMain function**:
   - Initializes the test environment by configuring it through flags, generating a new KubeConfig, and setting up the Kubernetes client.
   - Configures a new ServiceAccount, sets up the necessary roles, and assigns them for privilege-controlled testing.
   - Switches back to the original ServiceAccount after the test is finished.
   - **Modifiability**: The `TestMain` function is flexible and can be modified as needed. This structure is not mandatory and can be adjusted to suit specific test requirements.

2. **Setup Phase**:
   - Creates a new ServiceAccount and assigns the necessary ClusterRole.
   - Switches to the less-privileged ServiceAccount for test execution.

3. **Test Function**:
   - The test attempts to list cluster nodes and secrets using the less-privileged account, logging and handling any errors.

4. **Teardown Phase**:
   - Cleans up the resources created for the test (ServiceAccount, ClusterRole).
   - Switches back to the original privileged ServiceAccount.

#### Directory Structure for Tests

Tests should be organized under the `tests` directory, with a subdirectory for each test suite. Inside each test suite, a `testdata` directory will hold any necessary configuration or resource files (e.g., YAML files for ClusterRoles). For example, a privilege escalation test might have this structure:

Example structure:

```
tests/
  privilege_escalation/
    main_test.go
    testdata/
      node-list-cr.yaml
```

Each test suite can have multiple files but should include:
- `main_test.go`: Contains the `TestMain` function that sets up and tears down the environment.
- Other test files to store the actual tests and utility functions.

#### Executing the Test

To run the tests, you can use the following command:

```bash
$ go test -v ./path/to/test -args -sa-name "NAME" -sa-token "TOKEN" -cluster-endpoint "https://api.CLUSTER.DOMAIN:6443" -certificate-authority-data "BASE64_ENCODED_CERTIFICATE_AS_STRING"
```

Alternatively, you can see all available options using:

```bash
$ go test -v ./path/to/test --help
```

Or, the `config` package provides methods to configure these parameters programmatically. For example:

```go
...
a := config.New()
a.WithClusterName("my-cluster").
	WithClusterEndpoint("https://api.CLUSTER.DOMAIN:6443").
	WithCertificateAuthorityData("BASE64_ENCODED_CERTIFICATE").
	WithSAName("sa-name").
	WithSAToken("sa-token").
	WithNamespace("my-namespace")
path, err := NewKubeConfig(a)
...
```

The KubeConfig path can also be passed directly using the built-in flag `-kubeconfig` provided by the e2e-framework module.

#### `config` Package

The `config` package simplifies the process of creating KubeConfig files. It provides multiple functions to programmatically configure the attributes for a Kubernetes cluster, such as ServiceAccount name, token, cluster endpoint, and certificate data. These configurations can be passed as flags or directly set via the package’s functions.

Example functions include:

```go
func (a *AuthenticationAttr) WithClusterName(cn string) *AuthenticationAttr 

func (a *AuthenticationAttr) WithClusterEndpoint(ce string) *AuthenticationAttr

func (a *AuthenticationAttr) WithCertificateAuthorityData(cad string) *AuthenticationAttr

func (a *AuthenticationAttr) WithSAName(san string) *AuthenticationAttr

func (a *AuthenticationAttr) WithSAToken(sat string) *AuthenticationAttr
```

#### `escalation` Package

The `escalation` package provides functionality for switching between ServiceAccounts. This ensures that tests are executed with the least-privileged ServiceAccount necessary, minimizing security risks. Each test creates a specific ServiceAccount, assigns minimal permissions, and tests operations with the restricted account.

This package is designed to set up test environments with the minimum required permissions, ensuring that tests run with the least privilege necessary. By creating a dedicated ServiceAccount and assigning it a ClusterRole with only the specific permissions needed for each test, the security of the testing process is significantly enhanced. This approach minimizes the risk of over-privileged access, ensuring that tests are isolated and only capable of performing authorized operations.

This test structure ensures consistency across test cases, security in privilege escalation scenarios, and ease of configuration and execution across different environments.

For more in depth examples, see the tests directory for the various tests available.
