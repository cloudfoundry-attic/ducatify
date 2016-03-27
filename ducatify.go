package ducatify

import (
	"fmt"
	"strings"
)

type Transformer struct {
	ReleaseVersion   string
	DBPersistentDisk int
	DBResourcePool   string
	DBNetwork        string
}

func New() *Transformer {
	return &Transformer{
		ReleaseVersion:   "latest",
		DBPersistentDisk: 256,
		DBResourcePool:   "database_z1",
		DBNetwork:        "diego1",
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
	return nil
}

func dynRecover(err *error) {
	if e := recover(); e != nil {
		*err = fmt.Errorf("recovered: %+v", e)
	}
}

func (t *Transformer) addDucatiTemplate(manifest map[interface{}]interface{}, namePrefix string) (err error) {
	defer dynRecover(&err)

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
	defer dynRecover(&err)

	manifest["jobs"] = append(
		manifest["jobs"].([]interface{}),
		map[interface{}]interface{}{
			"name":            "ducati_db",
			"instances":       1,
			"persistent_disk": t.DBPersistentDisk,
			"resource_pool":   t.DBResourcePool,
			"networks": []interface{}{
				map[interface{}]interface{}{"name": t.DBNetwork},
			},
			"templates": []interface{}{
				map[interface{}]interface{}{"name": "postgres", "release": "ducati"},
				map[interface{}]interface{}{"name": "consul_agent", "release": "cf"},
			},
		})
	return nil
}

func (t *Transformer) updateReleases(manifest map[interface{}]interface{}) (err error) {
	defer dynRecover(&err)

	manifest["releases"] = append(
		manifest["releases"].([]interface{}),
		map[interface{}]interface{}{
			"name":    "ducati",
			"version": t.ReleaseVersion,
		})
	return nil
}
