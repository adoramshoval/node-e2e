package vm

import (
	"fmt"
	"strings"

	"node-e2e/utils"
	dv "node-e2e/utils/datavolume"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubev1 "kubevirt.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func GenerateVirtualMachine(v VM) *kubev1.VirtualMachine {
	vm := kubev1.VirtualMachine{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualMachine",
			APIVersion: "kubevirt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        v.VMName,
			Namespace:   v.Namespace,
			Labels:      v.Labels,
			Annotations: v.Annotations,
		},
		Spec: *generateVirtualMachineSpec(v.VMName, v.Namespace, v.VMSpec),
	}
	return &vm
}

func generateVirtualMachineSpec(vmname, ns string, vmspec VMSpec) *kubev1.VirtualMachineSpec {
	var dvTemplates []kubev1.DataVolumeTemplateSpec
	var counter uint = 1

	for _, d := range vmspec.DataVolumes {
		// Generate unique name for each DV
		var volName string = fmt.Sprintf("%s-%d", vmname, counter)
		// Generate DVs
		dvTemplates = append(dvTemplates, *dv.GenerateDataVolumeTemplateSpec(volName, d))
		bootOrder := counter
		// Create Volume structs for later kubev1.Volume and kubev1.Disk generation
		vmspec.VMISpec.Volumes = append(vmspec.VMISpec.Volumes, Volume{Name: volName, Type: DataVolume, Bootorder: &bootOrder})
		counter++
	}

	vms := kubev1.VirtualMachineSpec{
		Running:             &vmspec.Running,
		DataVolumeTemplates: dvTemplates,
		Template:            generateVirtualMachineInstanceTemplateSpec(vmname, ns, vmspec.VMISpec),
	}
	return &vms
}

func generateVirtualMachineInstanceTemplateSpec(vmname, ns string, vmispec VMISpec) *kubev1.VirtualMachineInstanceTemplateSpec {
	vmits := kubev1.VirtualMachineInstanceTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        vmname,
			Namespace:   ns,
			Labels:      vmispec.Labels,
			Annotations: vmispec.Annotations,
		},
		Spec: *generateVirtualmachineInstanceSpec(vmname, ns, vmispec),
	}
	return &vmits
}

func generateVirtualmachineInstanceSpec(vmname, ns string, vmispec VMISpec) *kubev1.VirtualMachineInstanceSpec {
	var networks []kubev1.Network
	var volumes []kubev1.Volume
	var nodeSelector map[string]string

	for _, net := range vmispec.Networks {
		// Generate kubev1.Network list
		networks = append(networks, generateNetwork(ns, net))
		// Generate kubev1.Interrace list based on the network list as they are based on the defined network
		vmispec.VMDomainSpec.interfaces = append(vmispec.VMDomainSpec.interfaces, generateInterface(net.Name, net.Type))
	}

	var defaultCloudInitNoCloud Volume = Volume{Name: "cloudinit", Type: CloudInitNoCloud}
	vmispec.Volumes = append(vmispec.Volumes, defaultCloudInitNoCloud)

	for _, vol := range vmispec.Volumes {
		if vol.Type == DataVolume {
			volumes = append(volumes, generateVolume(vol.Name))
			vmispec.VMDomainSpec.disks = append(vmispec.VMDomainSpec.disks, generateDisk(vol.Name, vol.Bootorder))
		} else if vol.Type == CloudInitNoCloud {
			volumes = append(volumes, generateCloudInitNoCloudVolume(vol.Name))
			vmispec.VMDomainSpec.disks = append(vmispec.VMDomainSpec.disks, generateDisk(vol.Name, nil))
		}
	}

	nodeSelector = make(map[string]string)
	if vmispec.AZ != nil {
		nodeSelector["topology.kubernetes.io/zone"] = *vmispec.AZ
	}
	if vmispec.NodeName != nil {
		nodeSelector["kubernetes.io/hostname"] = *vmispec.NodeName
	}

	spec := kubev1.VirtualMachineInstanceSpec{
		Domain:                        *generateDomainSpec(vmispec.VMDomainSpec),
		Hostname:                      vmname,
		Networks:                      networks,
		Affinity:                      vmispec.Affinity,
		Tolerations:                   vmispec.Tolerations,
		TopologySpreadConstraints:     vmispec.TopologySpreadConstraints,
		EvictionStrategy:              vmispec.EvictionStrategy,
		TerminationGracePeriodSeconds: vmispec.TerminationGracePeriodSeconds,
		Volumes:                       volumes,
		LivenessProbe:                 vmispec.LivenessProbe,
		ReadinessProbe:                vmispec.ReadinessProbe,
	}

	if len(nodeSelector) > 0 {
		spec.NodeSelector = nodeSelector
	}
	return &spec
}

func generateDomainSpec(domain VMDomainSpec) *kubev1.DomainSpec {
	domainSpec := kubev1.DomainSpec{
		CPU: &kubev1.CPU{
			Cores:   domain.Cores,
			Sockets: domain.Sockets,
			Threads: domain.Threads,
		},
		Devices: kubev1.Devices{
			Disks:                      domain.disks,
			Interfaces:                 domain.interfaces,
			NetworkInterfaceMultiQueue: func(b bool) *bool { return &b }(true),
			Rng:                        &kubev1.Rng{},
		},
		Machine: &kubev1.Machine{
			Type: "pc-q35-rhel8.6.0",
		},
		Resources: kubev1.ResourceRequirements{
			Limits:   *utils.GenerateResourceList(utils.GetCPULimitsFromRequests(domain.RequestsCPU), "", "", ""),
			Requests: *utils.GenerateResourceList(domain.RequestsCPU, domain.RequestsMemory, "", ""),
		},
	}

	return &domainSpec
}

// Basic Disk creation with disk device using vitio bus.
// There are many more options but for the purpose of testing this might be enough.
func generateDisk(devName string, bootorder *uint) kubev1.Disk {
	return kubev1.Disk{
		Name: devName,
		DiskDevice: kubev1.DiskDevice{
			Disk: &kubev1.DiskTarget{
				Bus: "virtio",
			},
		},
		BootOrder: bootorder, // Can be nil as it is a pointer
	}
}

// interfaceBindingMethod being "bridge" or otherwise default to Masquerade interface
func generateInterface(ifName string, interfaceBindingMethod NetworkType) kubev1.Interface {
	var iface kubev1.Interface
	if interfaceBindingMethod == BridgeNetwork {
		iface = *kubev1.DefaultBridgeNetworkInterface()
	} else {
		// Default to masquerade
		iface = *kubev1.DefaultMasqueradeNetworkInterface()
	}
	iface.Name = ifName
	iface.Model = "virtio"
	return iface
}

func generateVolume(name string) kubev1.Volume {
	volume := kubev1.Volume{
		Name: name,
		VolumeSource: kubev1.VolumeSource{
			DataVolume: &kubev1.DataVolumeSource{
				Name: name,
			},
		},
	}
	return volume
}

// This func allows only setting cloud-user password and SSH authorized keys
// and not the full functionality of cloud-init
func generateCloudInitNoCloudVolume(name string) kubev1.Volume {
	volume := kubev1.Volume{
		Name: name,
		VolumeSource: kubev1.VolumeSource{
			CloudInitNoCloud: &kubev1.CloudInitNoCloudSource{
				UserData: genUserData(generateRandPassword(3), ""),
			},
		},
	}
	return volume
}

func genUserData(password, sshKey string) string {
	return fmt.Sprintf(`#cloud-config
user: cloud-user
password: '%s'
chpasswd:
  expire: false
ssh_authorized_keys:
  - '%s'
`, password, sshKey)
}

func generateNetwork(ns string, net Network) kubev1.Network {
	network := kubev1.DefaultPodNetwork()
	if net.Type == BridgeNetwork {
		if net.NADName != nil {
			network.NetworkSource = kubev1.NetworkSource{
				Multus: &kubev1.MultusNetwork{
					// Format <namespace>/<nad-name> or <nad-name>
					NetworkName: fmt.Sprintf("%s/%s", ns, *net.NADName),
				},
			}
		} else {
			// If NADName is not provided but NetworkType "Bridge" is provided, network Name is assumed as the NAD name
			network.NetworkSource = kubev1.NetworkSource{
				Multus: &kubev1.MultusNetwork{
					NetworkName: fmt.Sprintf("%s/%s", ns, net.Name),
				},
			}
		}
	}
	network.Name = net.Name
	return *network
}

func generateRandPassword(parts int) string {
	var pass []string
	for i := 0; i < parts; i++ {
		pass = append(pass, envconf.RandomName("", 4))
	}
	return strings.Join(pass, "-")
}
