package selector

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GenMatchLabels(labels map[string]string) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: labels,
	}
}

func GenMatchExpressions(labelSelectorRequirements ...metav1.LabelSelectorRequirement) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchExpressions: labelSelectorRequirements[:],
	}
}

func GenLabelSelectorRequirement(key string, operator metav1.LabelSelectorOperator, values ...string) *metav1.LabelSelectorRequirement {
	return &metav1.LabelSelectorRequirement{
		Key:      key,
		Operator: operator,
		Values:   values[:],
	}
}
