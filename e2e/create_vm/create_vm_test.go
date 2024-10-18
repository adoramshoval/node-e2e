package createvm

import (
	"context"
	"fmt"
	vmconditions "node-e2e/utils/conditions"
	dv "node-e2e/utils/datavolume"
	"node-e2e/utils/vm"
	"testing"

	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubev1 "kubevirt.io/api/core/v1"

	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestVMCreateInteract(t *testing.T) {
	// Populate VM specification
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
			// Initialize the map if it's nil
			if testVM.ObjectMeta.Labels == nil {
				testVM.ObjectMeta.Labels = make(map[string]string)
			}
			if testVM.Spec.Template.ObjectMeta.Labels == nil {
				testVM.Spec.Template.ObjectMeta.Labels = make(map[string]string)
			}
			// Merge with already existing labels
			for key, value := range labels {
				testVM.ObjectMeta.Labels[key] = value
				testVM.Spec.Template.ObjectMeta.Labels[key] = value
			}
			return ctx
		}).
		Assess("Create a new VirtualMachine and wait for VirtualMachineInstance and pod to appear", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var objList []k8s.Object
			var timeoutM int64 = 2

			if err := c.Client().Resources(namespace).Create(ctx, testVM); err != nil {
				t.Fatal(err)
			}

			objList = []k8s.Object{
				&kubev1.VirtualMachine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      vmname,
						Namespace: namespace,
					},
				},
				&kubev1.VirtualMachineInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      vmname,
						Namespace: namespace,
					},
				},
			}

			// Wait for VM and VMI resources to get created
			for _, obj := range objList {
				err := wait.For(conditions.New(c.Client().Resources().WithNamespace(namespace)).ResourceMatch(obj, func(object k8s.Object) bool { return true }), wait.WithTimeout(time.Duration(timeoutM)*time.Minute))
				if err != nil {
					t.Fatal(err)
				}
			}
			// Wait for Pod resource
			var podList corev1.PodList
			// Pod should be labeled with "kubevirt.io/domain=vmname" as done in the setup phase
			// Since this name is randomly generated there should be only 1 pod with the same label
			err := wait.For(conditions.New(c.Client().Resources().WithNamespace(namespace)).ResourceListN(&podList, 1, resources.WithLabelSelector(fmt.Sprintf("kubevirt.io/domain=%s", vmname))), wait.WithTimeout(time.Duration(timeoutM)*time.Minute))
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("VirtualMachine, %s, VirtualMachineInstance, %s and Pod, %s were created successfully", vmname, vmname, podList.Items[0].GetName())
			return ctx
		}).
		Assess("VirtualMachine, VirtualMachineInstance and VirtLauncher Pod are Ready", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var timeoutM int64 = 3
			var podList corev1.PodList

			vm := kubev1.VirtualMachine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      vmname,
					Namespace: namespace,
				},
			}
			vmi := kubev1.VirtualMachineInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      vmname,
					Namespace: namespace,
				},
			}
			// Fetching pod list - this populates podList variable
			if err := c.Client().Resources(namespace).List(ctx, &podList, resources.WithLabelSelector(fmt.Sprintf("kubevirt.io/domain=%s", vmname))); err != nil {
				t.Fatal(err)
			}

			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podList.Items[0].GetName(),
					Namespace: namespace,
				},
			}

			// wait for VM to become Ready
			err := wait.For(conditions.New(c.Client().Resources().WithNamespace(namespace)).ResourceMatch(&vm, vmconditions.VMReady()), wait.WithTimeout(time.Minute*time.Duration(timeoutM)))
			if err != nil {
				t.Fatal(err)
			}
			// Wait for VMI to become Ready
			err = wait.For(conditions.New(c.Client().Resources().WithNamespace(namespace)).ResourceMatch(&vmi, vmconditions.VMIReady()), wait.WithTimeout(time.Minute*time.Duration(timeoutM)))
			if err != nil {
				t.Fatal(err)
			}
			// Wait for pod to become Ready
			err = wait.For(conditions.New(c.Client().Resources().WithNamespace(namespace)).PodReady(&pod), wait.WithTimeout(time.Minute*time.Duration(timeoutM)))
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("VirtualMachine, %s, VirtualMachineInstance, %s and Pod, %s are now in Ready condition!", vmname, vmname, podList.Items[0].GetName())

			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {

			return ctx
		}).Feature()

	// testsEnvironment is the one global that we rely on; it passes the context
	// and *envconf.Config to our feature.
	testsEnvironment.Test(t, feat)
}
