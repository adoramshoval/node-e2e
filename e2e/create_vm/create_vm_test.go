package createvm

import (
	"context"
	dv "node-e2e/utils/datavolume"
	"node-e2e/utils/vm"
	"testing"

	// "time"

	corev1 "k8s.io/api/core/v1"
	kubev1 "kubevirt.io/api/core/v1"

	// "sigs.k8s.io/e2e-framework/klient/k8s"
	// "sigs.k8s.io/e2e-framework/klient/k8s/resources"
	// "sigs.k8s.io/e2e-framework/klient/wait"
	// "sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestVMCreateInteract(t *testing.T) {
	vm1 := &vm.VM{
		VMName:    vmname,
		Namespace: namespace,
		VMSpec: vm.VMSpec{
			Running:     true,
			RunStrategy: kubev1.RunStrategyAlways,
			DataVolumes: []dv.DataVolumeData{
				{
					DVSource:         dv.GenerateDataVolumeSourcePVC("openshift-virtualization-os-images", osImagePVC),
					PVAccessMode:     corev1.ReadWriteMany,
					StorageRequests:  "15Gi",
					PVMode:           corev1.PersistentVolumeBlock,
					StorageClassName: "az-a",
				},
			},
			VMISpec: vm.VMISpec{
				NodeName: nil,
				AZ:       func(s string) *string { return &s }("az-a"),
				Networks: []vm.Network{
					{
						Name: "nic-0",
						Type: vm.MasqueradeNetwork,
					},
				},
				VMDomainSpec: vm.VMDomainSpec{
					RequestsCPU:    "250m",
					RequestsMemory: "2Gi",
					Cores:          1,
					Sockets:        1,
					Threads:        1,
				},
			},
		},
	}

	// Generate the kubev1.VirtualMachine struct
	testVM := vm.GenerateVirtualMachine(*vm1)

	feat := features.New("VM Creation and Interacion").
		WithLabel("type", "VM").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// Set test labels on the VM struct
			for key, value := range labels {
				testVM.ObjectMeta.Labels[key] = value
				testVM.Spec.Template.ObjectMeta.Labels[key] = value
			}
			return ctx
		}).
		Assess("Test creating a new VirtualMachine", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			if err := c.Client().Resources(namespace).Create(ctx, testVM); err != nil {
				t.Fatal(err)
			}
			return ctx
		}).Assess("VM is running and ready", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		// wait for VM to become running and ready
		// err := wait.For(conditions.New(c.Client().Resources()).ResourceMatch(&testVM, func(object k8s.Object) bool {
		// 	vmLatest := c.Client().Resources().Get(ctx, vmname, )
		// 	return float64(d.Status.ReadyReplicas)/float64(*d.Spec.Replicas) >= 0.50
		// }), wait.WithTimeout(time.Minute*2))
		// if err != nil {
		// 	t.Fatal(err)
		// }

		return ctx
	}).Feature()

	// testsEnvironment is the one global that we rely on; it passes the context
	// and *envconf.Config to our feature.
	testsEnvironment.Test(t, feat)
}
