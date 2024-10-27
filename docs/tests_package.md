# `tests` Package Documentation

The `tests` package offers functions to assist with setting up Kubernetes environments for testing with temporary `ServiceAccount`s, managing `ClusterRole` assignments, and performing account switches and rollbacks. This document provides an overview of each function and includes a sample `TestMain` setup for user reference.

## Functions Overview

### 1. **`StartWithServiceAccountFlags`**
```go
func StartWithServiceAccountFlags(namespace string) (env.Environment, *escalation.ServiceAccount, error)
```

This function creates a `KubeConfig` file, an `escalation.ServiceAccount` object, and an `env.Environment`, returning them for test execution. It accepts a namespace as a parameter, but this can be overridden by a namespace passed via flags.

- **Parameters**
  - `namespace`: The namespace in which the `ServiceAccount` will operate.

- **Returns**
  - `env.Environment`: The test environment configured with the specified namespace and `KubeConfig`.
  - `*escalation.ServiceAccount`: The `ServiceAccount` used for test execution.
  - `error`: An error message, if one occurs during the environment setup.

---

### 2. **`SetupWithAccountSwitch`**
```go
func SetupWithAccountSwitch(saName, namespace, crPath string) func(ctx context.Context, c *envconf.Config) (*escalation.ServiceAccount, context.Context, error)
```

This function creates a new `ServiceAccount` with the given name and namespace, assigns a specified `ClusterRole` to it, and switches to this new account, storing the original one for future rollback.

- **Parameters**
  - `saName`: The name of the `ServiceAccount` to create and switch to.
  - `namespace`: The namespace where the `ServiceAccount` will be created.
  - `crPath`: The file path to the `ClusterRole` to be assigned to the `ServiceAccount`.

- **Returns**
  - `*escalation.ServiceAccount`: The newly created `ServiceAccount`.
  - `context.Context`: Updated context with the new `ServiceAccount`.
  - `error`: An error if the creation or switch fails.

---

### 3. **`FinishWithAccountRollback`**
```go
func FinishWithAccountRollback(privilegedAcc, unprivAcc *escalation.ServiceAccount, crPath string) func(ctx context.Context, c *envconf.Config) (context.Context, error)
```

This function reverts the environment to the original, privileged `ServiceAccount`, then cleans up by deleting the temporary `ServiceAccount`, `ClusterRole`, and `ClusterRoleBinding` created during the setup phase.

- **Parameters**
  - `privilegedAcc`: The original, privileged `ServiceAccount` used prior to the account switch.
  - `unprivAcc`: The temporary `ServiceAccount` created during setup.
  - `crPath`: The file path to the `ClusterRole` for cleanup.

- **Returns**
  - `context.Context`: Updated context after cleanup.
  - `error`: An error if the rollback or cleanup fails.

---

## Example Usage

To set up and tear down the test environment with these functions, use the following `TestMain` example. Ensure you define the required constants and variables before executing the tests.

### `TestMain` Function

```go
import (
	"context"
	"fmt"
	"os"
	"testing"

	"node-e2e/utils/escalation"
	"node-e2e/utils/tests"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

const (
	namespace          string = "default"               // The namespace to be used in testing
	saName             string = "node-checker"          // The name of the ServiceAccount to be created
	crPath             string = "testdata/node-lister.yaml"  // Path to the ClusterRole yaml file
)

var (
	testsEnvironment env.Environment
	privAcc          *escalation.ServiceAccount
	newAcc           *escalation.ServiceAccount
)

func TestMain(m *testing.M) {
	e, a, err := tests.StartWithServiceAccountFlags(namespace)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	testsEnvironment = e
	privAcc = a

	testsEnvironment.Setup(func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		a, newCtx, err := tests.SetupWithAccountSwitch(saName, namespace, crPath)(ctx, c)
		ctx = newCtx
		if err != nil {
			fmt.Printf("Setup failure: %v", err)
			os.Exit(1)
		}
		newAcc = a

        // Additional setup logic can be added here if required

		return ctx, nil
	})
	testsEnvironment.Finish(func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		newCtx, err := tests.FinishWithAccountRollback(privAcc, newAcc, crPath)(ctx, c)
		ctx = newCtx
		if err != nil {
			return ctx, err
		}

        // Additional finish logic can be added here if required

		return ctx, nil
	})

	rc := testsEnvironment.Run(m)
	os.Exit(rc)
}
```

### Explanation of Variables
- **`namespace`**: Defines the namespace where the tests will create and manage resources.
- **`saName`**: Specifies the name of the `ServiceAccount` to create during the setup.
- **`crPath`**: Path to the YAML file defining the `ClusterRole` that will be assigned to the `ServiceAccount`.
- **`testsEnvironment`**: The main test environment created by `StartWithServiceAccountFlags`, used to configure the test lifecycle.
- **`privAcc` and `newAcc`**: Used to store the original, privileged `ServiceAccount` and the temporary `ServiceAccount` created during the setup.

### Running the Test

The `TestMain` function initializes and configures the test environment using the `StartWithServiceAccountFlags` function, which leverages user-provided flags for custom settings. This approach allows you to pass various Kubernetes cluster credentials and configuration parameters as command-line flags, streamlining test setup and enabling secure, temporary access to cluster resources. 

#### Key Flags

The `StartWithServiceAccountFlags` function accepts the following flags:

- `-sa-name`: Name of the `ServiceAccount` to use for test authentication.
- `-sa-token`: The token associated with the specified `ServiceAccount` for secure cluster access.
- `-cluster-endpoint`: The API server endpoint for the Kubernetes cluster.
- `-certificate-authority-data`: The base64-encoded string of the API server’s certificate authority data, required for secure TLS connections.
- `-dir-name`: Optional directory name for storing the generated `KubeConfig` file. By default, this file is saved in the user’s home directory but, optionally, a subdirectory can be passed.

#### Example Execution

To execute the test suite with a `ServiceAccount` and cluster-specific flags, use the following command:

```bash
go test -v ./e2e/test-directory -args \
-sa-name "NAME" \
-cluster-endpoint "https://api.medone-1.med.one:6443" \
-sa-token "TOKEN-STRING-AFTER-B64-DECODE" \
-certificate-authority-data "DATA-B64-ENCODED"
```

In this example:
- The `-v` flag enables verbose test output.
- The `./e2e/test-directory` specifies the directory containing your test files.
- Each `-args` flag is passed to `StartWithServiceAccountFlags` to configure the `KubeConfig`, cluster endpoint, and `ServiceAccount` for secure authentication.

#### How It Works
Once executed, `TestMain` initializes the `testsEnvironment` using the flags passed to `StartWithServiceAccountFlags`. The environment lifecycle is then managed through the following steps:
1. **Setup**: The `testsEnvironment.Setup` function applies the `SetupWithAccountSwitch`, creating and switching to a temporary `ServiceAccount` with appropriate roles.
2. **Test Execution**: The environment executes all tests within the specified directory, applying the setup configuration.
3. **Teardown**: The `testsEnvironment.Finish` function runs `FinishWithAccountRollback`, cleaning up all resources created in the setup and switching back to the original, privileged `ServiceAccount`.

This setup provides a controlled, reproducible test environment with secure, temporary elevated access to Kubernetes resources, ensuring both test isolation and security.