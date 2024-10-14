package utils

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// Help struct for comparing NodeSystemInfo without semantic values such as IDs
type nodeInfo struct {
	kernelVersion           string
	osImage                 string
	containerRuntimeVersion string
	kubeletVersion          string
	kubeProxyVersion        string
	operatingSystem         string
	arch                    string
}

var (
	nodeConditionTypes = []corev1.NodeConditionType{
		corev1.NodeMemoryPressure,
		corev1.NodeDiskPressure,
		corev1.NodePIDPressure,
		corev1.NodeNetworkUnavailable,
	}
)

func IsNodePerfectState(node *corev1.Node) (*corev1.NodeConditionType, bool) {
	condType, ok := nodePerfectCondition(&node.Status)
	return condType, ok
}

func nodePerfectCondition(node *corev1.NodeStatus) (*corev1.NodeConditionType, bool) {
	for _, cond := range node.Conditions {
		for _, condType := range nodeConditionTypes {
			if cond.Type == condType && cond.Status != corev1.ConditionFalse {
				return &cond.Type, false
			}
		}
		if cond.Type == corev1.NodeReady && cond.Status != corev1.ConditionTrue {
			return &cond.Type, false
		}
	}
	return nil, true
}

// This will compare two nodes' NodeSystemInfo and provide a list of the differences for display
func SystemInfoDifference(refNode *corev1.NodeSystemInfo, node *corev1.NodeSystemInfo) ([]string, bool) {
	var difList []string

	if populateNodeInfo(refNode) != populateNodeInfo(node) {
		if node.KernelVersion != refNode.KernelVersion {
			difList = append(difList, fmt.Sprintf("KernelVersion differs: %s vs %s", refNode.KernelVersion, node.KernelVersion))
		}
		if node.OSImage != refNode.OSImage {
			difList = append(difList, fmt.Sprintf("OSImage differs: %s vs %s", refNode.OSImage, node.OSImage))
		}
		if node.ContainerRuntimeVersion != refNode.ContainerRuntimeVersion {
			difList = append(difList, fmt.Sprintf("ContainerRuntimeVersion differs: %s vs %s", refNode.ContainerRuntimeVersion, node.ContainerRuntimeVersion))
		}
		if node.KubeletVersion != refNode.KubeletVersion {
			difList = append(difList, fmt.Sprintf("KubeletVersion differs: %s vs %s", refNode.KubeletVersion, node.KubeletVersion))
		}
		if node.OperatingSystem != refNode.OperatingSystem {
			difList = append(difList, fmt.Sprintf("OperatingSystem differs: %s vs %s", refNode.OperatingSystem, node.OperatingSystem))
		}
		if node.Architecture != refNode.Architecture {
			difList = append(difList, fmt.Sprintf("Architecture differs: %s vs %s", refNode.Architecture, node.Architecture))
		}
		return difList, false
	}

	return nil, true
}

func populateNodeInfo(node *corev1.NodeSystemInfo) nodeInfo {
	return nodeInfo{
		kernelVersion:           node.KernelVersion,
		osImage:                 node.OSImage,
		containerRuntimeVersion: node.ContainerRuntimeVersion,
		kubeletVersion:          node.KubeletVersion,
		kubeProxyVersion:        node.KubeProxyVersion,
		operatingSystem:         node.OperatingSystem,
		arch:                    node.Architecture,
	}
}
