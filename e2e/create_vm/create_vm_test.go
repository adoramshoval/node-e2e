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
	"k8s.io/apimachinery/pkg/types"
	kubev1 "kubevirt.io/api/core/v1"

	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestVMCreateInteract(t *testing.T) {
	var featName string = "VM Creation and Interacion"

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

	feat := features.New(featName).
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
		Assess("Create a new VirtualMachine and wait for VirtualMachineInstance and Pod to appear", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var objList []k8s.Object

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
				if err := wait.For(conditions.New(c.Client().Resources().WithNamespace(namespace)).ResourceMatch(obj, func(object k8s.Object) bool { return true }),
					wait.WithTimeout(time.Duration(pollTimeoutMinutes)*time.Minute),
					wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second)); err != nil {
					t.Fatal(err)
				}
			}
			// Wait for Pod resource
			var podList corev1.PodList
			// Pod should be labeled with "kubevirt.io/domain=vmname" as done in the setup phase
			// Since this name is randomly generated there should be only 1 pod with the same label
			if err := wait.For(conditions.New(c.Client().Resources().WithNamespace(namespace)).ResourceListN(&podList, 1, resources.WithLabelSelector(fmt.Sprintf("kubevirt.io/domain=%s", vmname))),
				wait.WithTimeout(time.Duration(pollTimeoutMinutes)*time.Minute),
				wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second)); err != nil {
				t.Fatal(err)
			}

			t.Logf("VirtualMachine, %s, VirtualMachineInstance, %s and Pod, %s were created successfully", vmname, vmname, podList.Items[0].GetName())
			return ctx
		}).
		Assess("VirtualMachine, VirtualMachineInstance and VirtLauncher Pod are Ready", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {

			resourcesFunc := getTestResources(fmt.Sprintf("kubevirt.io/domain=%s", vmname))
			vm, vmi, pod := resourcesFunc(ctx, t, c)

			// wait for VM to become Ready
			if err := wait.For(conditions.New(c.Client().Resources().WithNamespace(namespace)).ResourceMatch(vm, vmconditions.VMReady()),
				wait.WithTimeout(time.Minute*time.Duration(pollTimeoutMinutes)),
				wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second)); err != nil {
				t.Fatal(err)
			}
			// Wait for VMI to become Ready
			if err := wait.For(conditions.New(c.Client().Resources().WithNamespace(namespace)).ResourceMatch(vmi, vmconditions.VMIReady()),
				wait.WithTimeout(time.Minute*time.Duration(pollTimeoutMinutes)),
				wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second)); err != nil {
				t.Fatal(err)
			}
			// Wait for pod to become Ready
			if err := wait.For(conditions.New(c.Client().Resources().WithNamespace(namespace)).PodReady(pod),
				wait.WithTimeout(time.Minute*time.Duration(pollTimeoutMinutes)),
				wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second)); err != nil {
				t.Fatal(err)
			}

			t.Logf("VirtualMachine, %s, VirtualMachineInstance, %s and Pod, %s are now in Ready condition!", vmname, vmname, pod.GetName())

			return ctx
		}).
		Assess("Restart the VirtualMachine and wait for it to enter Ready state", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {

			// Fetch VM, VMI and Pod
			resourcesFunc := getTestResources(fmt.Sprintf("kubevirt.io/domain=%s", vmname))
			vm, vmi, pod := resourcesFunc(ctx, t, c)

			// Patch VM with running false to trigger a VM shutdown
			patchData := []byte(`{"spec": {"running":false}}`)
			if err := c.Client().Resources(namespace).Patch(ctx, vm, k8s.Patch{PatchType: types.MergePatchType, Data: patchData}); err != nil {
				t.Fatal(err)
			}
			t.Logf("VirtualMachine, %s, was triggered for a shutdown", vmname)

			// Wait for VirtualMachineInstance and Pod to be deleted
			if err := wait.For(conditions.New(c.Client().Resources(namespace)).ResourceDeleted(vmi),
				wait.WithTimeout(time.Minute*time.Duration(pollTimeoutMinutes)),
				wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second)); err != nil {
				t.Fatal(err)
			}
			if err := wait.For(conditions.New(c.Client().Resources(namespace)).ResourceDeleted(pod),
				wait.WithTimeout(time.Minute*time.Duration(pollTimeoutMinutes)),
				wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second)); err != nil {
				t.Fatal(err)
			}

			// Patch VM with running true to trigger a VM start
			patchData = []byte(`{"spec": {"running":true}}`)
			if err := c.Client().Resources(namespace).Patch(ctx, vm, k8s.Patch{PatchType: types.MergePatchType, Data: patchData}); err != nil {
				t.Fatal(err)
			}
			t.Logf("VirtualMachine, %s, was triggered to start up", vmname)

			vmi = &kubev1.VirtualMachineInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      vmname,
					Namespace: namespace,
				},
			}

			// Wait for VMI to get recreated
			if err := wait.For(conditions.New(c.Client().Resources().WithNamespace(namespace)).ResourceMatch(vmi, func(object k8s.Object) bool { return true }),
				wait.WithTimeout(time.Minute*time.Duration(pollTimeoutMinutes)),
				wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second)); err != nil {
				t.Fatal(err)
			}

			// Assuming that if the VMI was recreated the pod also, and therefore could be fetched
			_, _, pod = resourcesFunc(ctx, t, c)
			// Wait for VMI to become Ready
			if err := wait.For(conditions.New(c.Client().Resources().WithNamespace(namespace)).ResourceMatch(vmi, vmconditions.VMIReady()),
				wait.WithTimeout(time.Minute*time.Duration(pollTimeoutMinutes)),
				wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second)); err != nil {
				t.Fatal(err)
			}
			// Wait for pod to become Ready
			if err := wait.For(conditions.New(c.Client().Resources().WithNamespace(namespace)).PodReady(pod),
				wait.WithTimeout(time.Minute*time.Duration(pollTimeoutMinutes)),
				wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second)); err != nil {
				t.Fatal(err)
			}

			t.Logf("VirtualMachine, %s, was restarted successfully!", vmname)

			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var gracePeriodSeconds int64 = 30

			resourcesFunc := getTestResources(fmt.Sprintf("kubevirt.io/domain=%s", vmname))
			vm, vmi, pod := resourcesFunc(ctx, t, c)

			if err := c.Client().Resources(namespace).Delete(ctx, vm, resources.WithGracePeriod(time.Duration(gracePeriodSeconds)*time.Second)); err != nil {
				t.Fatal(err)
			}
			// Poll (pollTimeoutMinutes * 60 / pollIntervalSeconds) times before failing the test
			if err := wait.For(conditions.New(c.Client().Resources(namespace)).ResourceDeleted(vm),
				wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second),
				wait.WithTimeout(time.Duration(pollTimeoutMinutes)*time.Minute)); err != nil {
				t.Fatal(err)
			}
			// Making sure VirtualMachineInstance and VirtLauncher Pod were deleted as well
			if err := wait.For(conditions.New(c.Client().Resources(namespace)).ResourceDeleted(vmi), wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second)); err != nil {
				t.Fatal(err)
			}
			if err := wait.For(conditions.New(c.Client().Resources(namespace)).ResourceDeleted(pod), wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second)); err != nil {
				t.Fatal(err)
			}
			t.Logf("All resources have been deleted. %s test has finished successfully!", featName)

			return ctx
		}).Feature()

	// testsEnvironment is the one global that we rely on; it passes the context
	// and *envconf.Config to our feature.
	testsEnvironment.Test(t, feat)
}

// Helper function used in order to retrive the VM, VMI and Pod created during the test based on the label put on them
func getTestResources(label string) func(ctx context.Context, t *testing.T, c *envconf.Config) (*kubev1.VirtualMachine, *kubev1.VirtualMachineInstance, *corev1.Pod) {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) (*kubev1.VirtualMachine, *kubev1.VirtualMachineInstance, *corev1.Pod) {
		var vmList kubev1.VirtualMachineList
		var vmiList kubev1.VirtualMachineInstanceList
		var podList corev1.PodList

		// Fetching VM list - this populates vmList variable
		if err := c.Client().Resources(namespace).List(ctx, &vmList, resources.WithLabelSelector(label)); err != nil {
			t.Fatal(err)
		}

		vm := kubev1.VirtualMachine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      vmList.Items[0].GetName(),
				Namespace: namespace,
			},
		}

		// Fetching VMI list - this populates vmiList variable
		if err := c.Client().Resources(namespace).List(ctx, &vmiList, resources.WithLabelSelector(label)); err != nil {
			t.Fatal(err)
		}

		vmi := kubev1.VirtualMachineInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      vmiList.Items[0].GetName(),
				Namespace: namespace,
			},
		}

		// Fetching pod list - this populates podList variable
		if err := c.Client().Resources(namespace).List(ctx, &podList, resources.WithLabelSelector(label)); err != nil {
			t.Fatal(err)
		}

		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      podList.Items[0].GetName(),
				Namespace: namespace,
			},
		}

		return &vm, &vmi, &pod
	}
}
