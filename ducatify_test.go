package ducatify_test

import (
	"github.com/cloudfoundry-incubator/ducatify"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transform", func() {
	var (
		manifest    map[interface{}]interface{}
		transformer *ducatify.Transformer
	)

	BeforeEach(func() {
		transformer = ducatify.New()
		manifest = map[interface{}]interface{}{
			"releases": []interface{}{},
			"jobs": []interface{}{
				map[interface{}]interface{}{
					"name":      "cell_z1",
					"templates": []interface{}{},
				},
			},
		}
	})

	// trying to use new naming convention:
	// https://github.com/cloudfoundry/bosh-notes/blob/master/deployment-naming.md
	Describe("modifying cell instance groups", func() {
		BeforeEach(func() {
			manifest["jobs"] = []interface{}{
				map[interface{}]interface{}{
					"name":      "cell_z1",
					"instances": 3,
					"templates": []interface{}{
						map[interface{}]interface{}{"name": "some-template", "release": "some-release"},
					},
				},
				map[interface{}]interface{}{
					"name":      "cell_z2",
					"instances": 5,
					"templates": []interface{}{
						map[interface{}]interface{}{"name": "some-template", "release": "some-release"},
					},
				},
				map[interface{}]interface{}{
					"name":      "colocated_z3",
					"instances": 1,
					"templates": []interface{}{
						map[interface{}]interface{}{"name": "some-template", "release": "some-release"},
					},
				},
			}
		})

		It("colocates ducati template onto every cell instance group", func() {
			err := transformer.Transform(manifest)
			Expect(err).NotTo(HaveOccurred())
			jobs := manifest["jobs"].([]interface{})
			Expect(jobs[0]).To(Equal(map[interface{}]interface{}{
				"name":      "cell_z1",
				"instances": 3,
				"templates": []interface{}{
					map[interface{}]interface{}{"name": "some-template", "release": "some-release"},
					map[interface{}]interface{}{"name": "ducati", "release": "ducati"},
				},
			}))
			Expect(jobs[1]).To(Equal(map[interface{}]interface{}{
				"name":      "cell_z2",
				"instances": 5,
				"templates": []interface{}{
					map[interface{}]interface{}{"name": "some-template", "release": "some-release"},
					map[interface{}]interface{}{"name": "ducati", "release": "ducati"},
				},
			}))
		})
	})

	Describe("modifying the colocated instance", func() {
		BeforeEach(func() {
			manifest["jobs"] = []interface{}{
				map[interface{}]interface{}{
					"name":      "colocated_z3",
					"instances": 1,
					"templates": []interface{}{
						map[interface{}]interface{}{"name": "some-template", "release": "some-release"},
					},
				},
			}
		})
		It("colocates ducati template onto the 'colocated' job", func() {
			err := transformer.Transform(manifest)
			Expect(err).NotTo(HaveOccurred())
			jobs := manifest["jobs"].([]interface{})
			Expect(jobs[0]).To(Equal(map[interface{}]interface{}{
				"name":      "colocated_z3",
				"instances": 1,
				"templates": []interface{}{
					map[interface{}]interface{}{"name": "some-template", "release": "some-release"},
					map[interface{}]interface{}{"name": "ducati", "release": "ducati"},
				},
			}))
		})
	})

	Describe("adding new jobs", func() {
		It("adds the ducati_db job", func() {
			err := transformer.Transform(manifest)
			Expect(err).NotTo(HaveOccurred())
			jobs := manifest["jobs"].([]interface{})
			Expect(jobs).To(ContainElement(map[interface{}]interface{}{
				"name":            "ducati_db",
				"instances":       1,
				"persistent_disk": 256,
				"resource_pool":   "database_z1",
				"networks": []interface{}{
					map[interface{}]interface{}{"name": "diego1"},
				},
				"templates": []interface{}{
					map[interface{}]interface{}{"name": "postgres", "release": "ducati"},
					map[interface{}]interface{}{"name": "consul_agent", "release": "cf"},
				},
			}))
		})
	})

	Describe("updating releases", func() {
		BeforeEach(func() {
			manifest["releases"] = []interface{}{
				map[interface{}]interface{}{"name": "some-release", "version": "latest"},
				map[interface{}]interface{}{"name": "another-release", "version": "whatever"},
			}
		})

		It("adds the ducati release", func() {
			err := transformer.Transform(manifest)
			Expect(err).NotTo(HaveOccurred())

			Expect(manifest).To(HaveKey("releases"))
			releases := manifest["releases"]
			Expect(releases).To(ContainElement(
				map[interface{}]interface{}{"name": "some-release", "version": "latest"},
			))
			Expect(releases).To(ContainElement(
				map[interface{}]interface{}{"name": "another-release", "version": "whatever"},
			))
			Expect(releases).To(ContainElement(
				map[interface{}]interface{}{"name": "ducati", "version": "latest"},
			))
		})
	})

})
