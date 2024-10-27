package createvm

import (
	"context"
	"fmt"
	"os"
	"testing"

	"node-e2e/utils/escalation"
	"node-e2e/utils/tests"

	kubev1 "kubevirt.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

const (
	saName              string = "vm-creator"
	namespace           string = "default"
	vmNamePrefix        string = "node-e2e"
	osImagePVC          string = "rhel7-9-az-a"
	pollIntervalSeconds int64  = 10
	pollTimeoutMinutes  int64  = 5
	crPath              string = "testdata/vm-controller.yaml"
)

var (
	testsEnvironment env.Environment
	vmname           string = envconf.RandomName(vmNamePrefix, 13) // Generate a random VM name
	// Label VM and VMI
	labels map[string]string = map[string]string{
		"kubevirt.io/domain": vmname,
	}
	privAcc *escalation.ServiceAccount
	newAcc  *escalation.ServiceAccount
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

		// Add kubevirt.io to runtime scheme for later interaction with API groups it provides
		kubev1.AddToScheme(c.Client().Resources().GetScheme())

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
