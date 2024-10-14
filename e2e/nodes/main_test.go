package nodes_test

import (
	"context"
	"os"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

const (
	namespace    string        = "default"
	resourceType string        = "Nodes"
	fetchTimeout time.Duration = 30 * time.Second
)

var testsEnvironment = env.New()

func TestMain(m *testing.M) {
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
