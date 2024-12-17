// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and onMetal contributors
// SPDX-License-Identifier: Apache-2.0

package onmetal

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGcp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Suite")
}
