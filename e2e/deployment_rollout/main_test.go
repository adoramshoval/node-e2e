package createdaemonset_test

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

var (
	testsEnvironment env.Environment
	namespace        string            = "default"
	saName           string            = "dep-manager"
	crPath           string            = "testdata/dep-manager-cr.yaml"
	testLabels       map[string]string = map[string]string{
		"test": envconf.RandomName(saName, 20),
	}
	pollIntervalSeconds int64 = 10
	pollTimeoutMinutes  int64 = 2
	privAcc             *escalation.ServiceAccount
	newAcc              *escalation.ServiceAccount
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
		return ctx, nil
	})
	testsEnvironment.Finish(func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		newCtx, err := tests.FinishWithAccountRollback(privAcc, newAcc, crPath)(ctx, c)
		ctx = newCtx
		if err != nil {
			return ctx, err
		}
		return ctx, nil
	})
	os.Exit(testsEnvironment.Run(m))
}
