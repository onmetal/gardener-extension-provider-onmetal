// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and onMetal contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/controlplane"
	"github.com/gardener/gardener/extensions/pkg/controller/controlplane/genericactuator"
	"github.com/gardener/gardener/extensions/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/onmetal/gardener-extension-provider-onmetal/imagevector"
	"github.com/onmetal/gardener-extension-provider-onmetal/pkg/onmetal"
)

var (
	// DefaultAddOptions are the default AddOptions for AddToManager.
	DefaultAddOptions = AddOptions{}
)

// AddOptions are options to apply when adding the onmetal controlplane controller to the manager.
type AddOptions struct {
	// Controller are the controller.Options.
	Controller controller.Options
	// IgnoreOperationAnnotation specifies whether to ignore the operation annotation or not.
	IgnoreOperationAnnotation bool
	// WebhookServerNamespace is the namespace in which the webhook server runs.
	WebhookServerNamespace string
}

// AddToManagerWithOptions adds a controller with the given Options to the given manager.
// The opts.Reconciler is being set with a newly instantiated actuator.
func AddToManagerWithOptions(ctx context.Context, mgr manager.Manager, opts AddOptions) error {
	genericActuator, err := genericactuator.NewActuator(mgr,
		onmetal.ProviderName,
		secretConfigsFunc,
		shootAccessSecretsFunc,
		nil,
		nil,
		configChart,
		controlPlaneChart,
		controlPlaneShootChart,
		nil,
		storageClassChart,
		nil,
		NewValuesProvider(mgr),
		extensionscontroller.ChartRendererFactoryFunc(util.NewChartRendererForShoot),
		imagevector.ImageVector(),
		onmetal.CloudProviderConfigName,
		nil,
		opts.WebhookServerNamespace)

	if err != nil {
		return err
	}

	return controlplane.Add(ctx, mgr, controlplane.AddArgs{
		Actuator:          genericActuator,
		ControllerOptions: opts.Controller,
		Predicates:        controlplane.DefaultPredicates(ctx, mgr, opts.IgnoreOperationAnnotation),
		Type:              onmetal.Type,
	})
}

// AddToManager adds a controller with the default Options.
func AddToManager(ctx context.Context, mgr manager.Manager) error {
	return AddToManagerWithOptions(ctx, mgr, DefaultAddOptions)
}
