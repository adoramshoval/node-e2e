package conditions

import (
	corev1 "k8s.io/api/core/v1"
	kubev1 "kubevirt.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
)

func VMConditionMatch(conditionType kubev1.VirtualMachineConditionType, conditionStatus corev1.ConditionStatus) func(obj k8s.Object) bool {
	return func(obj k8s.Object) bool {
		vm, ok := obj.(*kubev1.VirtualMachine)
		if !ok {
			return false
		}

		for _, cond := range vm.Status.Conditions {
			if cond.Type == conditionType && cond.Status == conditionStatus {
				return true
			}
		}
		return false
	}
}

func VMIPhaseMatch(phase kubev1.VirtualMachineInstancePhase) func(obj k8s.Object) bool {
	return func(obj k8s.Object) bool {
		vmi, ok := obj.(*kubev1.VirtualMachineInstance)
		if !ok {
			return false
		}

		return vmi.Status.Phase == phase
	}
}

func VMIConditionMatch(conditionType kubev1.VirtualMachineInstanceConditionType, conditionStatus corev1.ConditionStatus) func(obj k8s.Object) bool {
	return func(obj k8s.Object) bool {
		vmi, ok := obj.(*kubev1.VirtualMachineInstance)
		if !ok {
			return false
		}

		for _, cond := range vmi.Status.Conditions {
			if cond.Type == conditionType && cond.Status == conditionStatus {
				return true
			}
		}
		return false
	}
}

func VMReady() func(obj k8s.Object) bool {
	return VMConditionMatch(kubev1.VirtualMachineReady, corev1.ConditionTrue)
}

func VMIReady() func(obj k8s.Object) bool {
	return VMIConditionMatch(kubev1.VirtualMachineInstanceReady, corev1.ConditionTrue)
}

func VMIRunning() func(obj k8s.Object) bool {
	return VMIPhaseMatch(kubev1.Running)
}
