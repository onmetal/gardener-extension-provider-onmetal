// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and onMetal contributors
// SPDX-License-Identifier: Apache-2.0

//go:generate sh -c "bash $GARDENER_HACK_DIR/generate-controller-registration.sh provider-onmetal . $(cat ../../VERSION) ../../example/controller-registration.yaml BackupBucket:onmetal BackupEntry:onmetal Bastion:onmetal ControlPlane:onmetal Infrastructure:onmetal Worker:onmetal"

// Package chart enables go:generate support for generating the correct controller registration.
package chart
