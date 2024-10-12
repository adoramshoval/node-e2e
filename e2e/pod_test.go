package e2e

import (
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"
	kubev1 "kubevirt.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

// The following shows an example of a simple
// test function that reaches out to the API server.
func TestAPICall(t *testing.T) {
	feat := features.New("API Feature").
		WithLabel("type", "API").
		Assess("test message", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var sas v1.ServiceAccountList
			var vmis kubev1.VirtualMachineInstanceList

			kubev1.AddToScheme(c.Client().Resources().GetScheme())

			if err := c.Client().Resources("core").List(ctx, &sas); err != nil {
				t.Error(err)
			}
			t.Logf("Got ServiceAccounts %v in namespace", len(sas.Items))
			if len(sas.Items) == 0 {
				t.Errorf("Expected >0 ServiceAccounts in core but got %v", len(sas.Items))
			}

			if err := c.Client().Resources("core").List(ctx, &vmis); err != nil {
				t.Error(err)
			}
			t.Logf("Got VMIs %v in namespace", len(vmis.Items))
			if len(vmis.Items) == 0 {
				t.Errorf("Expected >0 VMIs in core but got %v", len(vmis.Items))
			}
			return ctx
		}).Feature()

	// testsEnvironment is the one global that we rely on; it passes the context
	// and *envconf.Config to our feature.
	testsEnvironment.Test(t, feat)
}
