package utils

import (
	corev1 "k8s.io/api/core/v1"
)

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
