package e2e

import (
	"context"
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

var testsEnvironment = env.New()

func TestMain(m *testing.M) {
	// Setup the rand number source and a limit
	testsEnvironment.Setup(func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		client, err := config.NewClient()
		if err != nil {
			return ctx, err
		}

		// Assign the generated client to the Config's client attribute
		config.WithClient(client)

		return ctx, nil
	})

	// Don't forget to launch the package test
	os.Exit(testsEnvironment.Run(m))
}
