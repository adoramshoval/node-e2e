package nodes_test

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
	resourceType       string = "Nodes"
	namespace          string = "default"
	saName             string = "node-checker"
	crPath             string = "testdata/node-lister.yaml"
	pollTimeoutMinutes int64  = 2
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

	rc := testsEnvironment.Run(m)
	os.Exit(rc)
}
