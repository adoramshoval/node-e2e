package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubev1 "kubevirt.io/api/core/v1"
)

/*
   spec:
     domain:
       cpu:
         cores: 1
         sockets: 1
         threads: 1
       devices:
         disks:
           - disk:
               bus: virtio
             name: cloudinitdisk
           - bootOrder: 1
             disk:
               bus: virtio
             name: rhel7-9-8bwxw5
         interfaces:
           - macAddress: '02:f7:71:00:00:12'
             masquerade: {}
             model: virtio
             name: nic-0
         networkInterfaceMultiqueue: true
         rng: {}
       machine:
         type: pc-q35-rhel8.6.0
       resources:
         limits:
           cpu: '1'
         requests:
           cpu: 250m
           memory: 2Gi
*/

type VMData struct {
	Running     bool
	RunStrategy kubev1.VirtualMachineRunStrategy
}

type VMDomainSpec struct {
	RequestsCPU    string // Check how to multiply that by 4
	RequestsMemory string
	Cores          uint32
	Sockets        uint32
	Threads        uint32
	MachineType    string
}

func GenerateVirtualMachine(name, ns string) *kubev1.VirtualMachine {
	vm := kubev1.VirtualMachine{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualMachine",
			APIVersion: "kubevirt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: *generateVirtualMachineSpec(),
	}
	return &vm
}
func generateVirtualMachineSpec() *kubev1.VirtualMachineSpec

func generateVirtualMachineInstanceTemplateSpec() *kubev1.VirtualMachineInstanceTemplateSpec

func generateVirtualmachineInstanceSpec()

func generateDomainSpec(vmname, requestsCpu, requestsMem string, cores, sockets, threads uint32) *kubev1.DomainSpec {
	domainSpec := kubev1.DomainSpec{
		CPU: &kubev1.CPU{
			Cores:   cores,
			Sockets: sockets,
			Threads: threads,
		},
		Devices: kubev1.Devices{
			Disks: []kubev1.Disk{
				{
					Name: "cloudinitdisk",
					DiskDevice: kubev1.DiskDevice{
						Disk: &kubev1.DiskTarget{
							Bus: "virtio",
						},
					},
				},
				{
					Name: vmname,
					DiskDevice: kubev1.DiskDevice{
						Disk: &kubev1.DiskTarget{
							Bus: "virtio",
						},
					},
					BootOrder: func(i uint) *uint { return &i }(1),
				},
			},
			Interfaces: []kubev1.Interface{
				{
					Name:  "nic-0",
					Model: "virtio",
					InterfaceBindingMethod: kubev1.InterfaceBindingMethod{
						Masquerade: &kubev1.InterfaceMasquerade{},
					},
				},
			},
			NetworkInterfaceMultiQueue: func(b bool) *bool { return &b }(true),
			Rng:                        &kubev1.Rng{},
		},
		Machine: &kubev1.Machine{
			Type: "pc-q35-rhel8.6.0",
		},
		Resources: kubev1.ResourceRequirements{
			Limits:   *GenerateResourceList(getCPULimitsFromRequests(requestsCpu), "", "", ""),
			Requests: *GenerateResourceList(requestsCpu, requestsMem, "", ""),
		},
	}

	return &domainSpec
}
