// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and onMetal contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"

	"github.com/onmetal/gardener-extension-provider-onmetal/pkg/onmetal"
)

// ValidateCloudProviderSecret checks whether the given secret contains a valid onmetal service account.
func ValidateCloudProviderSecret(secret *corev1.Secret) error {
	if _, ok := secret.Data[onmetal.TokenFieldName]; !ok {
		return fmt.Errorf("missing field: %s in cloud provider secret", onmetal.TokenFieldName)
	}
	namespace, ok := secret.Data[onmetal.NamespaceFieldName]
	if !ok {
		return fmt.Errorf("missing field: %s in cloud provider secret", onmetal.NamespaceFieldName)
	}
	if _, ok := secret.Data[onmetal.UsernameFieldName]; !ok {
		return fmt.Errorf("missing field: %s in cloud provider secret", onmetal.UsernameFieldName)
	}
	errs := apivalidation.ValidateNamespaceName(string(namespace), false)
	if len(errs) > 0 {
		return fmt.Errorf("invalid field: %s in cloud provider secret", onmetal.NamespaceFieldName)
	}

	return nil
}
