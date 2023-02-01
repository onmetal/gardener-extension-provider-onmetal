// Copyright 2023 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package infrastructure

import (
	"encoding/json"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/onmetal/gardener-extension-provider-onmetal/pkg/apis/onmetal/v1alpha1"
	"github.com/onmetal/gardener-extension-provider-onmetal/pkg/onmetal"
	v1alpha12 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	testutils "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("Infrastructure Reconcile", func() {
	ctx := testutils.SetupContext()
	ns := SetupTest(ctx)

	It("should create a network, natgateway and prefix for a given infrastructure configuration", func() {
		By("getting the cluster object")
		cluster, err := extensionscontroller.GetCluster(ctx, k8sClient, ns.Name)
		Expect(err).NotTo(HaveOccurred())

		providerConfig := v1alpha1.InfrastructureConfig{}
		providerConfigJson, err := json.Marshal(providerConfig)
		Expect(err).NotTo(HaveOccurred())

		By("creating an infrastructure configuration")
		infra := &extensionsv1alpha1.Infrastructure{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      "my-infra",
				Annotations: map[string]string{
					constants.GardenerOperation: constants.GardenerOperationReconcile,
				},
			},
			Spec: extensionsv1alpha1.InfrastructureSpec{
				DefaultSpec: extensionsv1alpha1.DefaultSpec{
					Type:           onmetal.Type,
					ProviderConfig: &runtime.RawExtension{Raw: providerConfigJson},
				},
				Region: "foo",
				SecretRef: corev1.SecretReference{
					Name: "my-infra-creds",
				},
			},
		}
		Expect(k8sClient.Create(ctx, infra)).Should(Succeed())

		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, client.ObjectKeyFromObject(infra), infra)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(infra.Status.LastOperation).NotTo(BeNil())
		}).Should(Succeed())

		By("expecting a network being created")
		network := &networkingv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      generateResourceNameFromCluster(cluster),
			},
		}

		Eventually(Object(network)).Should(SatisfyAll(
			HaveField("ObjectMeta.OwnerReferences", ContainElement(metav1.OwnerReference{
				APIVersion:         extensionsv1alpha1.SchemeGroupVersion.String(),
				Kind:               "Infrastructure",
				Name:               infra.Name,
				UID:                infra.UID,
				Controller:         pointer.Bool(true),
				BlockOwnerDeletion: pointer.Bool(true),
			})),
		))

		By("expecting a nat gateway being created")
		natGateway := &networkingv1alpha1.NATGateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      generateResourceNameFromCluster(cluster),
			},
		}

		Eventually(Object(natGateway)).Should(SatisfyAll(
			HaveField("ObjectMeta.OwnerReferences", ContainElement(metav1.OwnerReference{
				APIVersion:         extensionsv1alpha1.SchemeGroupVersion.String(),
				Kind:               "Infrastructure",
				Name:               infra.Name,
				UID:                infra.UID,
				Controller:         pointer.Bool(true),
				BlockOwnerDeletion: pointer.Bool(true),
			})),
			HaveField("Spec.Type", networkingv1alpha1.NATGatewayTypePublic),
			HaveField("Spec.IPFamilies", []corev1.IPFamily{corev1.IPv4Protocol}),
			HaveField("Spec.NetworkRef", corev1.LocalObjectReference{
				Name: network.Name,
			}),
			HaveField("Spec.NetworkInterfaceSelector", &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      workerNameKey,
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"foo", "bar"},
					},
				},
			}),
		))

		By("expecting a nat gateway being created")
		prefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      generateResourceNameFromCluster(cluster),
			},
		}

		Eventually(Object(prefix)).Should(SatisfyAll(
			HaveField("ObjectMeta.OwnerReferences", ContainElement(metav1.OwnerReference{
				APIVersion:         extensionsv1alpha1.SchemeGroupVersion.String(),
				Kind:               "Infrastructure",
				Name:               infra.Name,
				UID:                infra.UID,
				Controller:         pointer.Bool(true),
				BlockOwnerDeletion: pointer.Bool(true),
			})),
			HaveField("Spec.IPFamily", corev1.IPv4Protocol),
			HaveField("Spec.Prefix", v1alpha12.MustParseNewIPPrefix("10.0.0.0/24")),
		))

		By("ensuring that the infrastructure state contains the correct refs")
		providerStatus := map[string]interface{}{
			"networkRef": map[string]interface{}{
				"name": network.Name,
				"uid":  network.UID,
			},
			"natGatewayRef": map[string]interface{}{
				"name": natGateway.Name,
				"uid":  natGateway.UID,
			},
			"prefixRef": map[string]interface{}{
				"name": prefix.Name,
				"uid":  prefix.UID,
			},
		}
		providerStatusJSON, err := json.Marshal(providerStatus)
		Expect(err).NotTo(HaveOccurred())
		Eventually(Object(infra)).Should(SatisfyAll(
			HaveField("Status.ProviderStatus", &runtime.RawExtension{Raw: providerStatusJSON}),
		))
	})
})
