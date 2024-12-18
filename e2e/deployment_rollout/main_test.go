package deployment_rollout

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
	namespace          string = "default"
	saName             string = "dep-manager"
	crPath             string = "testdata/dep-manager-cr.yaml"
	testImage          string = "quay.med.one:8443/openshift/ubi8/ubi"
	deploymentReplicas int32  = 3
)

var (
	testsEnvironment env.Environment
	workloadName     string            = envconf.RandomName(saName, 20)
	testLabels       map[string]string = map[string]string{
		"test": workloadName,
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
