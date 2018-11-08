package dpbuilder_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	artifacts "github.com/kun-lun/artifacts/pkg/apis"
	. "github.com/kun-lun/deployment-producer/dpbuilder"
	testinfra "github.com/kun-lun/test-infra/pkg/apis"
)

var _ = Describe("DpBuilder", func() {
	var (
		m *artifacts.Manifest
	)
	BeforeEach(func() {
		testInfra := testinfra.TestInfra{}
		m = testInfra.BuildSampleManifest()
	})
	Describe("Produce", func() {
		Context("Everything OK", func() {
			It("should produce deployments and hosts correctly", func() {
				dpProducer := DeploymentBuilder{}
				x, y, err := dpProducer.Produce(*m)
				Expect(err).To(BeNil())
				x_str, err := json.Marshal(x)
				Expect(err).To(BeNil())
				println(string(x_str))
				y_str, err := json.Marshal(y)
				Expect(err).To(BeNil())
				println(string(y_str))
			})
		})
	})
})
