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
			"properties": map[interface{}]interface{}{
				"something": "whatever",
				"garden": map[interface{}]interface{}{
					"a_thing": "a_value",
				},
			},
			"networks": []interface{}{
				map[interface{}]interface{}{
					"name": "diego1",
					"subnets": []interface{}{
						map[interface{}]interface{}{
							"static": []interface{}{"10.10.5.10 - 10.10.5.63"},
						},
					},
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
					map[interface{}]interface{}{
						"name":       "diego1",
						"static_ips": []interface{}{"10.10.5.10"},
					},
				},
				"templates": []interface{}{
					map[interface{}]interface{}{"name": "postgres", "release": "ducati"},
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

	Describe("adding garden properties", func() {
		It("sets the network plugin properties", func() {
			err := transformer.Transform(manifest)
			Expect(err).NotTo(HaveOccurred())

			Expect(manifest["properties"]).To(HaveKeyWithValue("garden",
				map[interface{}]interface{}{
					"a_thing": "a_value",
					"shared_mounts": []string{
						"/var/vcap/data/ducati/container-netns",
					},
					"network_plugin": "/var/vcap/packages/ducati/bin/guardian-cni-adapter",
					"network_plugin_extra_args": []string{
						"--configFile=/var/vcap/jobs/ducati/config/adapter.json",
					},
				}))
		})
	})

	Describe("adding ducati properties", func() {
		It("adds properties for ducati", func() {
			err := transformer.Transform(manifest)
			Expect(err).NotTo(HaveOccurred())

			Expect(manifest["properties"]).To(HaveKeyWithValue("ducati",
				map[interface{}]interface{}{
					"daemon": map[interface{}]interface{}{
						"database": map[interface{}]interface{}{
							"username": "ducati_daemon",
							"password": "some-password",
							"name":     "ducati",
							"ssl_mode": "disable",
							"host":     "10.10.5.10",
							"port":     5432,
						},
					},
					"database": map[interface{}]interface{}{
						"db_scheme": "postgres",
						"port":      5432,
						"databases": []interface{}{
							map[interface{}]interface{}{
								"name": "ducati", "tag": "whatever",
							},
						},
						"roles": []interface{}{
							map[interface{}]interface{}{
								"name":     "ducati_daemon",
								"password": "some-password",
								"tag":      "admin",
							},
						},
					},
				},
			))
		})

		Context("when there are errors", func() {
			Context("finding networks key", func() {
				It("returns reported error", func() {
					delete(manifest, "networks")

					err := transformer.Transform(manifest)
					Expect(err).To(MatchError("recovered: interface conversion: interface is nil, not []interface {}"))
				})
			})

			Context("finding subnets key in networks", func() {
				It("returns reported error", func() {
					networks := manifest["networks"].([]interface{})[0].(map[interface{}]interface{})
					delete(networks, "subnets")

					err := transformer.Transform(manifest)
					Expect(err).To(MatchError("recovered: interface conversion: interface is nil, not []interface {}"))
				})
			})

			Context("finding static key in subnets", func() {
				It("returns reported error", func() {
					subnets := manifest["networks"].([]interface{})[0].(map[interface{}]interface{})["subnets"].([]interface{})[0].(map[interface{}]interface{})
					delete(subnets, "static")

					err := transformer.Transform(manifest)
					Expect(err).To(MatchError("recovered: interface conversion: interface is nil, not []interface {}"))
				})
			})

			Context("parsing static ip range", func() {
				It("returns reported error", func() {
					subnets := manifest["networks"].([]interface{})[0].(map[interface{}]interface{})["subnets"].([]interface{})[0].(map[interface{}]interface{})
					subnets["static"] = []interface{}{""}

					err := transformer.Transform(manifest)
					Expect(err).To(MatchError("could not parse static ip range from "))
				})
			})
		})
	})
})
