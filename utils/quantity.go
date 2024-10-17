package utils

import (
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
)

const (
	requestsToLimitsRation int64 = 4
)

func parseQuantity(value string) *resource.Quantity {
	if value == "" {
		return nil // No value provided, omit the attribute
	}
	parsedQuantity, err := resource.ParseQuantity(value)
	if err != nil {
		return nil // Parsing failed, omit the attribute
	}
	return &parsedQuantity
}

func GenerateResourceList(cpu, memory, storage, ephStorage string) *corev1.ResourceList {
	var resourceList corev1.ResourceList

	resourceList = make(corev1.ResourceList)
	if parsedCpu, err := resource.ParseQuantity(cpu); err == nil {
		resourceList[corev1.ResourceCPU] = parsedCpu
	}
	if parsedMemory, err := resource.ParseQuantity(memory); err == nil {
		resourceList[corev1.ResourceMemory] = parsedMemory
	}
	if parsedStorage, err := resource.ParseQuantity(storage); err == nil {
		resourceList[corev1.ResourceStorage] = parsedStorage
	}
	if parsedEphStorage, err := resource.ParseQuantity(ephStorage); err == nil {
		resourceList[corev1.ResourceEphemeralStorage] = parsedEphStorage
	}
	return &resourceList
}

func multiplyQuantity(q *resource.Quantity, factor int64) *resource.Quantity {
	// Get the raw value
	originalValue := q.Value()

	// Multiply the value
	multipliedValue := originalValue * factor

	// Create a new quantity with the same format (BinarySI or DecimalSI)
	return resource.NewQuantity(multipliedValue, q.Format)
}

// Usually the ratio between CPU requests and CPU limits is 1:4
func GetCPULimitsFromRequests(requests string) string {
	var q resource.Quantity
	var limits string

	// Generate a zero value mili cores quantity in case the provided requests value isn't parsable
	q = *resource.NewMilliQuantity(0, resource.DecimalSI)
	if parsed, err := resource.ParseQuantity(requests); err == nil {
		limits = multiplyQuantity(&parsed, requestsToLimitsRation).String()
	} else {
		limits = q.String()
	}

	return limits
}
