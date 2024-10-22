- [ ] `Start` function which will:
```
    import (
        "context"
        "fmt"
        "os"
        "testing"

        "node-e2e/utils/config"

        "sigs.k8s.io/e2e-framework/pkg/env"
        "sigs.k8s.io/e2e-framework/pkg/envconf"
    ...
    ...
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
```
- [ ] Provide a skeleton for a test
- [ ] Helper functions for privilege changes
    - [ ] Create ServiceAccount and store its token
    - [ ] Assign the token to the existing `envconf.Config.Client().RESTConfig()`
    - [ ] Crteate a ClusterRole and assign it to the SA
    - [ ] 