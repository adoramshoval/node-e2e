package utils

import (
	resource "k8s.io/apimachinery/pkg/api/resource"
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
