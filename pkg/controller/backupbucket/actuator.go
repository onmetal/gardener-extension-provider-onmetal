// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and onMetal contributors
// SPDX-License-Identifier: Apache-2.0

package backupbucket

import (
	"context"
	"fmt"

	"github.com/gardener/gardener/extensions/pkg/controller/backupbucket"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	controllerconfig "github.com/onmetal/gardener-extension-provider-onmetal/pkg/apis/config"
	"github.com/onmetal/gardener-extension-provider-onmetal/pkg/onmetal"
)

type actuator struct {
	backupBucketConfig *controllerconfig.BackupBucketConfig
	client             client.Client
}

func newActuator(mgr manager.Manager, backupBucketConfig *controllerconfig.BackupBucketConfig) backupbucket.Actuator {
	return &actuator{
		client:             mgr.GetClient(),
		backupBucketConfig: backupBucketConfig,
	}
}

// Reconcile implements backupbucket.Actuator
func (a *actuator) Reconcile(ctx context.Context, log logr.Logger, backupBucket *extensionsv1alpha1.BackupBucket) error {
	log.V(2).Info("Reconciling BackupBucket")

	onmetalClient, namespace, err := onmetal.GetOnmetalClientAndNamespaceFromSecretRef(ctx, a.client, &backupBucket.Spec.SecretRef)
	if err != nil {
		return fmt.Errorf("failed to get onmetal client and namespace from cloudprovider secret: %w", err)
	}

	// If the generated secret in the backupbucket status not exists that means
	// no backupbucket exists, and it needs to be created.
	if backupBucket.Status.GeneratedSecretRef == nil {
		if err := validateConfiguration(a.backupBucketConfig); err != nil {
			return fmt.Errorf("failed to validate configuration: %w", err)
		}

		if err := a.ensureBackupBucket(ctx, namespace, onmetalClient, backupBucket); err != nil {
			return fmt.Errorf("failed to ensure backupbucket: %w", err)
		}
	}
	log.V(2).Info("Reconciled BackupBucket")
	return nil
}

func (a *actuator) Delete(ctx context.Context, log logr.Logger, backupBucket *extensionsv1alpha1.BackupBucket) error {
	log.V(2).Info("Deleting BackupBucket")
	onmetalClient, namespace, err := onmetal.GetOnmetalClientAndNamespaceFromSecretRef(ctx, a.client, &backupBucket.Spec.SecretRef)
	if err != nil {
		return fmt.Errorf("failed to get onmetal client and namespace from cloudprovider secret: %w", err)
	}

	bucket := &storagev1alpha1.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backupBucket.Name,
			Namespace: namespace,
		},
	}
	if err = onmetalClient.Delete(ctx, bucket); err != nil {
		return fmt.Errorf("failed to delete backup bucket: %v", err)
	}

	log.V(2).Info("Deleted BackupBucket")
	return nil
}
