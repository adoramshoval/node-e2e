package vm

import (
	dv "node-e2e/utils/datavolume"

	corev1 "k8s.io/api/core/v1"
	kubev1 "kubevirt.io/api/core/v1"
)

type VMDomainSpec struct {
	RequestsCPU    string
	RequestsMemory string
	Cores          uint32
	Sockets        uint32
	Threads        uint32
	disks          []kubev1.Disk
	interfaces     []kubev1.Interface
}

type VMISpec struct {
	VMDomainSpec VMDomainSpec
	Labels       map[string]string
	Annotations  map[string]string
	// Allows node or AZ selector only
	NodeName *string
	AZ       *string
	// Should create a generate func if not a complex struct
	Affinity *corev1.Affinity
	// Should create a generate func for each toleration if not a complex struct
	Tolerations []corev1.Toleration
	// Should create a generate func for each topologyspreadconstraint if not a complex struct
	TopologySpreadConstraints     []corev1.TopologySpreadConstraint
	EvictionStrategy              *kubev1.EvictionStrategy
	TerminationGracePeriodSeconds *int64
	// Should create a generate func if not a complex struct
	LivenessProbe *kubev1.Probe
	// Should create a generate func if not a complex struct
	ReadinessProbe *kubev1.Probe
	// Allowing only DataVolume, CloudInitNoCloud
	// - GenerateVolume
	// - GenerateCloudInitNoCloudVolume
	Volumes  []Volume
	Networks []Network
}

type VMSpec struct {
	VMISpec     VMISpec
	Running     bool
	RunStrategy kubev1.VirtualMachineRunStrategy
	DataVolumes []dv.DataVolumeData
}

type VM struct {
	VMName      string
	Namespace   string
	VMSpec      VMSpec
	Labels      map[string]string
	Annotations map[string]string
}

type Network struct {
	Name    string
	Type    NetworkType
	NADName *string
}

type NetworkType string

const (
	BridgeNetwork     NetworkType = "Bridge"
	MasqueradeNetwork NetworkType = "Masquerade"
)

type Volume struct {
	Name      string
	Type      VolumeType
	Bootorder *uint
}

type VolumeType string

const (
	DataVolume       VolumeType = "DataVolume"
	CloudInitNoCloud VolumeType = "CloudInitNoCloud"
)
