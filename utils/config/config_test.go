package config

import (
	"context"
	"fmt"
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

const (
	namespace string = "default"
)

var testsEnvironment env.Environment

func TestMain(m *testing.M) {
	// create config from flags (always in TestMain or init handler of the package before calling envconf.NewFromFlags())
	// This is needed in order to initilize flags provided by the e2e-framework module
	cfg, err := envconf.NewFromFlags()
	if err != nil {
		fmt.Printf("failed to build envconf from flags: %s", err)
		os.Exit(1)
	}
	testsEnvironment = env.NewWithConfig(cfg)

	testsEnvironment.Setup(func(ctx context.Context, config *envconf.Config) (context.Context, error) {

		return ctx, nil
	})

	// Don't forget to launch the package test
	os.Exit(testsEnvironment.Run(m))
}

func TestKubeConfigFromFlags(t *testing.T) {
	feat := features.New("KubeConfig from flags").
		WithLabel("type", "Config").
		Assess("Test creating a KubeConfig file from provided flags", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			a := New()
			a.WithNamespace(namespace)
			if err := NewKubeConfigFromFlags(a); err != nil {
				t.Fatal(err)
			}
			return ctx
		}).Feature()

	testsEnvironment.Test(t, feat)
}

func TestKubeConfigFromStruct(t *testing.T) {
	feat := features.New("KubeConfig from struct").
		WithLabel("type", "Config").
		Assess("Test creating a KubeConfig file from provided struct attributes", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			a := New()
			a.WithSAName("test-sa")
			a.WithSAToken("123")
			a.WithClusterEndpoint("https://api.medone-0.med.one:6443")
			a.WithClusterName("medone-0")
			a.WithNamespace(namespace)
			if err := NewKubeConfig(a); err != nil {
				t.Fatal(err)
			}
			return ctx
		}).Feature()

	testsEnvironment.Test(t, feat)
}
