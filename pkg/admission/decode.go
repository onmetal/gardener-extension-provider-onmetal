// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and onMetal contributors
// SPDX-License-Identifier: Apache-2.0

package admission

import (
	"github.com/gardener/gardener/extensions/pkg/util"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/onmetal/gardener-extension-provider-onmetal/pkg/apis/onmetal"
)

// DecodeControlPlaneConfig decodes the `ControlPlaneConfig` from the given `RawExtension`.
func DecodeControlPlaneConfig(decoder runtime.Decoder, cp *runtime.RawExtension) (*onmetal.ControlPlaneConfig, error) {
	controlPlaneConfig := &onmetal.ControlPlaneConfig{}
	if err := util.Decode(decoder, cp.Raw, controlPlaneConfig); err != nil {
		return nil, err
	}

	return controlPlaneConfig, nil
}

// DecodeInfrastructureConfig decodes the `InfrastructureConfig` from the given `RawExtension`.
func DecodeInfrastructureConfig(decoder runtime.Decoder, infra *runtime.RawExtension) (*onmetal.InfrastructureConfig, error) {
	infraConfig := &onmetal.InfrastructureConfig{}
	if err := util.Decode(decoder, infra.Raw, infraConfig); err != nil {
		return nil, err
	}

	return infraConfig, nil
}
