// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and onMetal contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	apisonmetal "github.com/onmetal/gardener-extension-provider-onmetal/pkg/apis/onmetal"
	"github.com/onmetal/gardener-extension-provider-onmetal/pkg/onmetal"
)

// ValidateInfrastructureConfig validates a InfrastructureConfig object.
func ValidateInfrastructureConfig(infra *apisonmetal.InfrastructureConfig, nodesCIDR, podsCIDR, servicesCIDR *string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if infra.NetworkRef != nil {
		for _, msg := range apivalidation.NameIsDNSLabel(infra.NetworkRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("networkRef").Child("name"), infra.NetworkRef.Name, msg))
		}
	}
	if infra.NATPortsPerNetworkInterface != nil && ValidatePowerOfTwo(*infra.NATPortsPerNetworkInterface) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("natPortsPerNetworkInterface"), infra.NATPortsPerNetworkInterface, "natPortsPerNetworkInterface must be a power of two."))
	}
	if infra.NATPortsPerNetworkInterface != nil && *infra.NATPortsPerNetworkInterface > onmetal.MaxAvailableNATPortsPerNetworkInterface {
		allErrs = append(allErrs, field.Invalid(field.NewPath("natPortsPerNetworkInterface"), infra.NATPortsPerNetworkInterface, "natPortsPerNetworkInterface can not be greater than max available NATPorts."))
	}

	return allErrs
}

// ValidateInfrastructureConfigUpdate validates a InfrastructureConfig object.
func ValidateInfrastructureConfigUpdate(oldConfig, newConfig *apisonmetal.InfrastructureConfig, fldPath *field.Path) field.ErrorList {
	var (
		allErrs = field.ErrorList{}
	)
	allErrs = append(allErrs, apivalidation.ValidateImmutableField(newConfig.NetworkRef, oldConfig.NetworkRef, fldPath.Child("networkRef"))...)

	return allErrs
}

func ValidatePowerOfTwo(value int32) bool {
	// Compare the binary representation of the given positive integer with its predecessor, e.g. '11011' (27) and '11010' (26).
	// They will share (at least) the leading '1' resulting in the union of them representing a number greater than zero, unless the given one is a power of two.
	return value <= 0 || value&(value-1) != 0
}
