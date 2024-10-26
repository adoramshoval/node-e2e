package tests

import (
	"context"
	"fmt"
	"path/filepath"

	"node-e2e/utils/config"
	"node-e2e/utils/escalation"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

// This will create a KubeConfig file, escalation.ServiceAccount object and env.Environment and return them
func StartWithServiceAccountFlags(namespace string) (env.Environment, *escalation.ServiceAccount, error) {
	var te env.Environment
	var acc *escalation.ServiceAccount

	// Build Environment configuration from provided flags to allow tests filtering and other capabilities
	cfg, err := envconf.NewFromFlags()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build environment configuration from flags: %s", err)
	}

	// Create a new AuthenticationAttr object
	a := config.New()
	// Set namespace for later KubeConfig generation - ns passed with a flag will over write parameter
	if cfg.Namespace() == "" {
		a.WithNamespace(namespace)
	} else {
		a.WithNamespace(cfg.Namespace())
	}

	// Generate a new KubeConfig from passed flags and store the result KC path
	kcPath, err := config.NewKubeConfigFromFlags(a)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create a new KubeConfig file: %v", err)
	}

	// store the currently used ServiceAccount object
	acc = escalation.New(a.GetServiceAccountName(), a.GetServiceAccountNamespace(), a.GetServiceAccountToken())

	// Set the config's KubeConfig
	cfg = cfg.WithKubeconfigFile(kcPath)

	// Finally create the Environment
	te = env.NewWithConfig(cfg)

	// Create a new klient.Client using the previously set KubeConfig
	cfg.NewClient()

	return te, acc, nil
}

// This will automatically search for a KubeConfig and create an env.Environment using it.
// Then, it will create escalation.ServiceAccount using the BearerToken of the create *rest.Config
func StartWithAutoResolve() (env.Environment, *escalation.ServiceAccount, error) {
	var te env.Environment
	var acc *escalation.ServiceAccount

	cfg := envconf.New()
	if _, err := cfg.NewClient(); err != nil {
		return nil, nil, fmt.Errorf("failed to create a new client: %v", err)
	}
	te = env.NewWithConfig(cfg)
	// IMPORTANT: cfg.Client().RESTConfig().BearerToken might be empty based on the authentication method
	// Therefore, this method might not always be valid for account switching
	acc = escalation.New("", "", cfg.Client().RESTConfig().BearerToken)

	return te, acc, nil
}

// Create an env.Environment using a -kubeconfig flag passed with a file path to a KubeConfig file.
// Then, it will create escalation.ServiceAccount using the BearerToken of the create *rest.Config
func StartWithKubeConfigAsFlag() (env.Environment, *escalation.ServiceAccount, error) {
	var te env.Environment
	var acc *escalation.ServiceAccount

	// Build Environment configuration from provided flags to allow tests filtering and other capabilities
	cfg, err := envconf.NewFromFlags()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build environment configuration from flags: %s", err)
	}

	if cfg.KubeconfigFile() == "" {
		return nil, nil, fmt.Errorf("no KubeConfig file path was passed")
	}

	// Finally create the Environment
	te = env.NewWithConfig(cfg)

	// Create a new klient.Client using the previously set KubeConfig
	cfg.NewClient()

	// IMPORTANT: cfg.Client().RESTConfig().BearerToken might be empty based on the authentication method
	// Therefore, this method might not always be valid for account switching
	acc = escalation.New("", "", cfg.Client().RESTConfig().BearerToken)

	return te, acc, nil
}

// This will create and switch to a newly created ServiceAccount. Then, it will create a ClusterRole
// using a specified file path and bind the CR to the ServiceAccount. It will return the escalation.ServiceAccount
// for later Rollback to original SA
func SetupWithAccountSwitch(saName, namespace, crPath string) func(ctx context.Context, c *envconf.Config) (*escalation.ServiceAccount, context.Context, error) {
	return func(ctx context.Context, c *envconf.Config) (*escalation.ServiceAccount, context.Context, error) {
		var newAcc *escalation.ServiceAccount

		// Create a new ServiceAccount with the provided name and namespace
		s, err := escalation.NewServiceAccount(saName, namespace)(ctx, c)
		if err != nil {
			return nil, ctx, err
		}
		newAcc = s

		// Get the absulute path to the ClusterRole to be created
		absCRPath, err := filepath.Abs(crPath)
		if err != nil {
			return nil, ctx, err
		}
		fmt.Print(absCRPath)
		// Assign the ClusterRole to the newly create ServiceAccount using the current, privileged, account
		if err := newAcc.AssignClusterRole(absCRPath)(ctx, c); err != nil {
			return nil, ctx, err
		}

		// Switch to the new ServiceAccount and store the old one
		_, err = escalation.SwitchAccount(newAcc)(ctx, c)
		if err != nil {
			return nil, ctx, err
		}

		return newAcc, ctx, nil
	}
}

// This will switch to a given escalation.ServiceAccount and delete previously created ServiceAccount, ClusterRole and Binding
// created during the setup phase using a provided file path to the ClusterRole
func FinishWithAccountRollback(privilegedAcc, unprivAcc *escalation.ServiceAccount, crPath string) func(ctx context.Context, c *envconf.Config) (context.Context, error) {
	return func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		// Switch to the privileged account and store the unprivileged one
		_, err := escalation.SwitchAccount(privilegedAcc)(ctx, c)
		if err != nil {
			return ctx, err
		}

		// Get the absulute path to the ClusterRole to be delete
		absCRPath, err := filepath.Abs(crPath)
		if err != nil {
			return ctx, err
		}

		// Attempt to clean up ServiceAccount, ClusterRole and ClusterRoleBinding
		if err := unprivAcc.CleanUp(absCRPath)(ctx, c); err != nil {
			return ctx, err
		}
		return ctx, nil
	}
}
