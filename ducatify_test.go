package ducatify_test

import (
	"github.com/cloudfoundry-incubator/ducatify"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transform", func() {
	var (
		manifest    map[string]interface{}
		transformer *ducatify.Transformer
	)

	BeforeEach(func() {
		transformer = ducatify.New()
		manifest = map[string]interface{}{
			"releases": []interface{}{},
			"jobs":     []interface{}{},
		}
	})

	Describe("modifying the jobs", func() {
		It("adds the ducati_db job", func() {
			err := transformer.Transform(manifest)
			Expect(err).NotTo(HaveOccurred())
			jobs := manifest["jobs"].([]interface{})
			Expect(jobs).To(ContainElement(map[string]interface{}{
				"name":            "ducati_db",
				"instances":       1,
				"persistent_disk": 256,
				"resource_pool":   "database_z1",
				"networks": []interface{}{
					map[string]interface{}{"name": "diego1"},
				},
				"templates": []interface{}{
					map[string]interface{}{"name": "postgres", "release": "ducati"},
					map[string]interface{}{"name": "consul_agent", "release": "cf"},
				},
			}))
		})
	})

	Describe("updating releases", func() {
		BeforeEach(func() {
			manifest["releases"] = []interface{}{
				map[string]interface{}{"name": "some-release", "version": "latest"},
				map[string]interface{}{"name": "another-release", "version": "whatever"},
			}
		})

		It("adds the ducati release", func() {
			err := transformer.Transform(manifest)
			Expect(err).NotTo(HaveOccurred())

			Expect(manifest).To(HaveKey("releases"))
			releases := manifest["releases"]
			Expect(releases).To(ContainElement(
				map[string]interface{}{"name": "some-release", "version": "latest"},
			))
			Expect(releases).To(ContainElement(
				map[string]interface{}{"name": "another-release", "version": "whatever"},
			))
			Expect(releases).To(ContainElement(
				map[string]interface{}{"name": "ducati", "version": "latest"},
			))
		})
	})

})
