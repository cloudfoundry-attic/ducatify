package ducatify

import (
	"errors"
	"fmt"
	"strings"
)

type Transformer struct {
	ReleaseVersion               string
	DBPersistentDisk             int
	DBResourcePool               string
	DBNetwork                    string
	GardenSharedMounts           []string
	GardenNetworkPlugin          string
	GardenNetworkPluginExtraArgs []string
	DBName                       string
	DBUsername                   string
	DBPassword                   string
	DBSSLMode                    string
	NsyncNetworkID               string
}

func New() *Transformer {
	return &Transformer{
		ReleaseVersion:   "latest",
		DBPersistentDisk: 256,
		DBResourcePool:   "database_z1",
		DBNetwork:        "diego1",

		DBName:     "ducati",
		DBUsername: "ducati_daemon",
		DBPassword: "some-password",
		DBSSLMode:  "disable",

		GardenSharedMounts:           []string{"/var/vcap/data/ducati/container-netns"},
		GardenNetworkPlugin:          "/var/vcap/packages/ducati/bin/guardian-cni-adapter",
		GardenNetworkPluginExtraArgs: []string{"--configFile=/var/vcap/jobs/ducati/config/adapter.json"},

		NsyncNetworkID: "ducati-overlay",
	}
}

func (t *Transformer) Transform(manifest map[interface{}]interface{}) error {
	err := t.updateReleases(manifest)
	if err != nil {
		return fmt.Errorf("updating releases: %s", err)
	}

	err = t.addDucatiDBJob(manifest)
	if err != nil {
		return fmt.Errorf("adding ducati_db job: %s", err)
	}

	err = t.addDucatiTemplate(manifest, "cell_z")
	if err != nil {
		return fmt.Errorf("adding ducati template to cells: %s", err)
	}

	err = t.addDucatiTemplate(manifest, "colocated_z")
	if err != nil {
		return fmt.Errorf("adding ducati template to colocated vm: %s", err)
	}

	err = t.addGardenProperties(manifest)
	if err != nil {
		return fmt.Errorf("adding garden properties: %s", err)
	}

	err = t.addNsyncProperties(manifest)
	if err != nil {
		return fmt.Errorf("adding nsync properties: %s", err)
	}

	err = t.addDucatiProperties(manifest)
	if err != nil {
		return fmt.Errorf("adding garden properties: %s", err)
	}
	return nil
}

func dynRecover(context string, err *error) {
	if e := recover(); e != nil {
		*err = fmt.Errorf("%s: %+v", context, e)
	}
}

func (t *Transformer) addDucatiTemplate(manifest map[interface{}]interface{}, namePrefix string) (err error) {
	defer dynRecover("add ducati template to "+namePrefix, &err)

	for _, jobVal := range manifest["jobs"].([]interface{}) {
		nameVal, err := getElement(jobVal, "name")
		if err != nil {
			return err
		}
		if !strings.HasPrefix(nameVal.(string), namePrefix) {
			continue
		}

		templates, err := getElement(jobVal, "templates")
		if err != nil {
			return err
		}
		templates = append(templates.([]interface{}),
			map[interface{}]interface{}{"name": "ducati", "release": "ducati"},
		)

		err = setElement(jobVal, "templates", templates)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Transformer) addDucatiDBJob(manifest map[interface{}]interface{}) (err error) {
	defer dynRecover("add ducati db job", &err)

	ducatiDBJob := map[interface{}]interface{}{
		"name":            "ducati_db",
		"instances":       1,
		"persistent_disk": t.DBPersistentDisk,
		"resource_pool":   t.DBResourcePool,
		"networks": []interface{}{
			map[interface{}]interface{}{
				"name": t.DBNetwork,
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
	}

	oldJobs := manifest["jobs"].([]interface{})
	newJobs := []interface{}{}
	for _, job := range oldJobs {
		newJobs = append(newJobs, job)
		if job.(map[interface{}]interface{})["name"] == "database_z1" {
			newJobs = append(newJobs, ducatiDBJob)
		}
	}
	if len(newJobs) == len(oldJobs) {
		return errors.New("database_z1 job not found, don't know where to put the ducati_db job")
	}

	manifest["jobs"] = newJobs

	return nil
}

func (t *Transformer) updateReleases(manifest map[interface{}]interface{}) (err error) {
	defer dynRecover("update releases", &err)

	manifest["releases"] = append(
		manifest["releases"].([]interface{}),
		map[interface{}]interface{}{
			"name":    "ducati",
			"version": t.ReleaseVersion,
		})
	return nil
}

func (t *Transformer) addGardenProperties(manifest map[interface{}]interface{}) (err error) {
	defer dynRecover("add garden properties", &err)
	gardenProps := manifest["properties"].(map[interface{}]interface{})["garden"].(map[interface{}]interface{})
	gardenProps["network_plugin"] = t.GardenNetworkPlugin
	gardenProps["network_plugin_extra_args"] = t.GardenNetworkPluginExtraArgs
	gardenProps["shared_mounts"] = t.GardenSharedMounts
	return nil
}

func (t *Transformer) addNsyncProperties(manifest map[interface{}]interface{}) (err error) {
	defer dynRecover("add nsync properties", &err)
	diegoProps := manifest["properties"].(map[interface{}]interface{})["diego"].(map[interface{}]interface{})
	nsyncProps := diegoProps["nsync"].(map[interface{}]interface{})
	nsyncProps["network_id"] = t.NsyncNetworkID
	return nil
}

func (t *Transformer) addDucatiProperties(manifest map[interface{}]interface{}) (err error) {
	defer dynRecover("add ducati properties", &err)

	props := manifest["properties"].(map[interface{}]interface{})
	props["ducati"] = map[interface{}]interface{}{
		"daemon": map[interface{}]interface{}{
			"database": map[interface{}]interface{}{
				"username": t.DBUsername,
				"password": t.DBPassword,
				"name":     t.DBName,
				"ssl_mode": t.DBSSLMode,
				"host":     "ducati-db.service.cf.internal",
				"port":     5432,
			},
		},
		"database": map[interface{}]interface{}{
			"db_scheme": "postgres",
			"port":      5432,
			"databases": []interface{}{
				map[interface{}]interface{}{
					"name": t.DBName, "tag": "whatever",
				},
			},
			"roles": []interface{}{
				map[interface{}]interface{}{
					"name":     t.DBUsername,
					"password": t.DBPassword,
					"tag":      "admin",
				},
			},
		},
	}

	return nil
}
