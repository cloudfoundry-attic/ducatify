package ducatify_test

import (
	"github.com/cloudfoundry-incubator/ducatify"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transform", func() {
	var (
		manifest            map[interface{}]interface{}
		acceptanceJobConfig map[interface{}]interface{}
		transformer         *ducatify.Transformer
		systemDomain        string
	)

	BeforeEach(func() {
		transformer = ducatify.New()
		systemDomain = "some.system.domain"
		manifest = map[interface{}]interface{}{
			"releases": []interface{}{},
			"jobs": []interface{}{
				map[interface{}]interface{}{
					"name":      "database_z1",
					"templates": []interface{}{},
				},
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
				"diego": map[interface{}]interface{}{
					"nsync": map[interface{}]interface{}{
						"bbs": "bbs_addr",
					},
					"route_emitter": map[interface{}]interface{}{
						"nats": map[interface{}]interface{}{
							"some-key": "some-value",
						},
					},
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

		acceptanceJobConfig = map[interface{}]interface{}{
			"api":                 "api.systemdomain.mycf.example.com",
			"admin_password":      "some-admin-password",
			"admin_user":          "some-admin-user",
			"apps_domain":         "appsdomain.mycf.example.com",
			"skip_ssl_validation": "true",
		}
	})

	// trying to use new naming convention:
	// https://github.com/cloudfoundry/bosh-notes/blob/master/deployment-naming.md
	Describe("modifying cell instance groups", func() {
		BeforeEach(func() {
			manifest["jobs"] = []interface{}{
				map[interface{}]interface{}{
					"name":      "database_z1",
					"instances": 1,
				},
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
				map[interface{}]interface{}{
					"name":      "cc_bridge_z1",
					"instances": 2,
					"templates": []interface{}{
						map[interface{}]interface{}{"name": "some-template", "release": "some-release"},
					},
				},
			}
		})

		It("colocates connet job template onto every cc_bridge instance group", func() {
			err := transformer.Transform(manifest, acceptanceJobConfig, systemDomain)
			Expect(err).NotTo(HaveOccurred())
			jobs := manifest["jobs"].([]interface{})
			Expect(jobs[5]).To(Equal(map[interface{}]interface{}{
				"name":      "cc_bridge_z1",
				"instances": 2,
				"properties": map[interface{}]interface{}{
					"route_registrar": map[interface{}]interface{}{
						"routes": []interface{}{
							map[interface{}]interface{}{
								"name":                  "connet",
								"registration_interval": "20s",
								"port":                  4002,
								"uris":                  []string{"connet.some.system.domain"},
							},
						},
					},
				},
				"templates": []interface{}{
					map[interface{}]interface{}{"name": "some-template", "release": "some-release"},
					map[interface{}]interface{}{"name": "connet", "release": "ducati"},
					map[interface{}]interface{}{"name": "route_registrar", "release": "cf"},
				},
			}))
		})

		It("colocates ducati template onto every cell instance group", func() {
			err := transformer.Transform(manifest, acceptanceJobConfig, systemDomain)
			Expect(err).NotTo(HaveOccurred())
			jobs := manifest["jobs"].([]interface{})
			Expect(jobs[2]).To(Equal(map[interface{}]interface{}{
				"name":      "cell_z1",
				"instances": 3,
				"templates": []interface{}{
					map[interface{}]interface{}{"name": "some-template", "release": "some-release"},
					map[interface{}]interface{}{"name": "ducati", "release": "ducati"},
				},
			}))
			Expect(jobs[3]).To(Equal(map[interface{}]interface{}{
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
					"name":      "database_z1",
					"instances": 1,
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

		It("colocates ducati template onto the 'colocated' job", func() {
			err := transformer.Transform(manifest, acceptanceJobConfig, systemDomain)
			Expect(err).NotTo(HaveOccurred())
			jobs := manifest["jobs"].([]interface{})
			Expect(jobs[2]).To(Equal(map[interface{}]interface{}{
				"name":      "colocated_z3",
				"instances": 1,
				"templates": []interface{}{
					map[interface{}]interface{}{"name": "some-template", "release": "some-release"},
					map[interface{}]interface{}{"name": "ducati", "release": "ducati"},
					map[interface{}]interface{}{"name": "connet", "release": "ducati"},
					map[interface{}]interface{}{"name": "route_registrar", "release": "cf"},
				},
			}))
		})
	})

	Describe("adding new jobs", func() {
		It("adds the ducati_db job", func() {
			err := transformer.Transform(manifest, acceptanceJobConfig, systemDomain)
			Expect(err).NotTo(HaveOccurred())
			jobs := manifest["jobs"].([]interface{})
			Expect(jobs).To(ContainElement(map[interface{}]interface{}{
				"name":            "ducati_db",
				"instances":       1,
				"persistent_disk": 256,
				"resource_pool":   "database_z1",
				"networks": []interface{}{
					map[interface{}]interface{}{
						"name": "diego1",
					},
				},
				"templates": []interface{}{
					map[interface{}]interface{}{"name": "postgres", "release": "ducati"},
					map[interface{}]interface{}{"name": "consul_agent", "release": "cf"},
				},
				"properties": map[interface{}]interface{}{
					"consul": map[interface{}]interface{}{
						"agent": map[interface{}]interface{}{
							"services": map[interface{}]interface{}{
								"ducati-db": map[interface{}]interface{}{
									"name": "ducati-db",
									"check": map[interface{}]interface{}{
										"script":   "/bin/true",
										"interval": "5s",
									},
								},
							},
						},
					},
				},
			}))
		})

		It("adds the ducati acceptance test job", func() {
			err := transformer.Transform(manifest, acceptanceJobConfig, systemDomain)
			Expect(err).NotTo(HaveOccurred())
			jobs := manifest["jobs"].([]interface{})
			Expect(jobs).To(ContainElement(map[interface{}]interface{}{
				"name":          "ducati-acceptance",
				"instances":     1,
				"lifecycle":     "errand",
				"resource_pool": "database_z1",
				"networks": []interface{}{
					map[interface{}]interface{}{
						"name": "diego1",
					},
				},
				"templates": []interface{}{
					map[interface{}]interface{}{"name": "acceptance-with-cf", "release": "ducati"},
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
			err := transformer.Transform(manifest, acceptanceJobConfig, systemDomain)
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
			err := transformer.Transform(manifest, acceptanceJobConfig, systemDomain)
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

	Describe("adding nsync properties", func() {
		It("sets the nsync network id", func() {
			err := transformer.Transform(manifest, acceptanceJobConfig, systemDomain)
			Expect(err).NotTo(HaveOccurred())

			Expect(manifest["properties"]).To(HaveKeyWithValue("diego",
				map[interface{}]interface{}{
					"nsync": map[interface{}]interface{}{
						"bbs":        "bbs_addr",
						"network_id": "ducati-overlay",
					},
					"route_emitter": map[interface{}]interface{}{
						"nats": map[interface{}]interface{}{
							"some-key": "some-value",
						},
					},
				},
			))
		})
	})

	Describe("adding ducati properties", func() {
		It("adds properties for ducati", func() {
			err := transformer.Transform(manifest, acceptanceJobConfig, systemDomain)
			Expect(err).NotTo(HaveOccurred())

			Expect(manifest["properties"]).To(HaveKeyWithValue("ducati",
				map[interface{}]interface{}{
					"daemon": map[interface{}]interface{}{
						"database": map[interface{}]interface{}{
							"username": "ducati_daemon",
							"password": "some-password",
							"name":     "ducati",
							"ssl_mode": "disable",
							"host":     "ducati-db.service.cf.internal",
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
	})

	Describe("adding acceptance-with-cf properties", func() {
		It("adds properties for acceptance with ducati", func() {
			err := transformer.Transform(manifest, acceptanceJobConfig, systemDomain)
			Expect(err).NotTo(HaveOccurred())

			Expect(manifest["properties"]).To(HaveKeyWithValue("acceptance-with-cf", acceptanceJobConfig))
		})
	})
})
