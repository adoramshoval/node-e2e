package utils

import (
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
)

const (
	requestsToLimitsRation int64 = 4
)

func GenerateResourceList(cpu, memory, storage, ephStorage string) *corev1.ResourceList {
	var resourceList = make(corev1.ResourceList)

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
	// Get the value in millicores to maintain precision
	originalMilliValue := q.MilliValue()

	// Multiply the value
	multipliedMilliValue := originalMilliValue * factor

	// Create a new quantity from the millicores result
	return resource.NewMilliQuantity(multipliedMilliValue, q.Format)
}

// Usually the ratio between CPU requests and CPU limits is 1:4
func GetCPULimitsFromRequests(requests string) string {
	// Parse the CPU requests string to a Quantity
	parsed, err := resource.ParseQuantity(requests)
	if err != nil {
		// If parsing fails, return a default 0 millicores Quantity as string
		return resource.NewMilliQuantity(0, resource.DecimalSI).String()
	}

	// Multiply the parsed Quantity by the ratio
	limits := multiplyQuantity(&parsed, requestsToLimitsRation)

	// Return the new limits Quantity as string
	return limits.String()
}
