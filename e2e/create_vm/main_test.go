package createvm

import (
	"context"
	"os"
	"testing"

	kubev1 "kubevirt.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

const (
	namespace    string = "core"
	vmNamePrefix string = "node-e2e"
	osImagePVC   string = "rhel7-9-az-a"
)

var testsEnvironment = env.New()
var vmname string = envconf.RandomName(vmNamePrefix, 13) // Generate a random VM name
// Label VM and VMI
var labels map[string]string = map[string]string{
	"kubevirt.io/domain": vmname,
}

func TestMain(m *testing.M) {
	testsEnvironment.Setup(func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		client, err := config.NewClient()
		if err != nil {
			return ctx, err
		}

		// Assign the generated client to the Config's client attribute
		config.WithClient(client)

		// Add kubevirt.io v1 to runtime scheme
		kubev1.AddToScheme(config.Client().Resources().GetScheme())

		return ctx, nil
	})

	// Don't forget to launch the package test
	os.Exit(testsEnvironment.Run(m))
}
