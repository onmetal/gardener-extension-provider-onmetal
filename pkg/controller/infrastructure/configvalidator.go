// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and onMetal contributors
// SPDX-License-Identifier: Apache-2.0

package infrastructure

import (
	"context"
	"fmt"

	"github.com/gardener/gardener/extensions/pkg/controller/infrastructure"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/onmetal/gardener-extension-provider-onmetal/pkg/apis/onmetal/helper"
	"github.com/onmetal/gardener-extension-provider-onmetal/pkg/onmetal"
)

// configValidator implements ConfigValidator for onmetal infrastructure resources.
type configValidator struct {
	client client.Client
	logger logr.Logger
}

// NewConfigValidator creates a new ConfigValidator.
func NewConfigValidator(client client.Client, logger logr.Logger) infrastructure.ConfigValidator {
	return &configValidator{
		client: client,
		logger: logger.WithName("onmetal-infrastructure-config-validator"),
	}
}

// Validate validates the provider config of the given infrastructure resource with the cloud provider.
func (c *configValidator) Validate(ctx context.Context, infra *extensionsv1alpha1.Infrastructure) field.ErrorList {
	allErrs := field.ErrorList{}

	if infra == nil || infra.Spec.ProviderConfig == nil {
		return allErrs
	}

	// Get provider config from the infrastructure resource
	config, err := helper.InfrastructureConfigFromInfrastructure(infra)
	if err != nil {
		allErrs = append(allErrs, field.InternalError(nil, err))
		return allErrs
	}

	// check wether a networkRef is set
	if config.NetworkRef == nil {
		return allErrs
	}

	// get onmetal credentials from infrastructure config
	onmetalClient, namespace, err := onmetal.GetOnmetalClientAndNamespaceFromCloudProviderSecret(ctx, c.client, infra.Namespace)
	if err != nil {
		allErrs = append(allErrs, field.InternalError(nil, fmt.Errorf("could not get onmetal client and namespace: %w", err)))
		return allErrs
	}

	// ensure that the referenced network exists
	network := &networkingv1alpha1.Network{}
	if err := onmetalClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: config.NetworkRef.Name}, network); err != nil {
		if apierrors.IsNotFound(err) {
			allErrs = append(allErrs, field.NotFound(field.NewPath("networkRef"), fmt.Errorf("could not find onmetal network %s: %w", client.ObjectKeyFromObject(network), err)))
			return allErrs
		}
		allErrs = append(allErrs, field.InternalError(field.NewPath("networkRef"), fmt.Errorf("failed to get onmetal network %s: %w", client.ObjectKeyFromObject(network), err)))
	}

	return allErrs
}
