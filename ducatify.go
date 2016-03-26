package ducatify

import (
	"errors"
	"fmt"
	"strings"
)

type Transformer struct {
	AppendToSlice    func(slice interface{}, toAppend interface{}) ([]interface{}, error)
	ReleaseVersion   string
	DBPersistentDisk int
	DBResourcePool   string
	DBNetwork        string
}

func New() *Transformer {
	return &Transformer{
		AppendToSlice:    appendToSlice,
		ReleaseVersion:   "latest",
		DBPersistentDisk: 256,
		DBResourcePool:   "database_z1",
		DBNetwork:        "diego1",
	}
}

func (t *Transformer) Transform(manifest map[string]interface{}) error {
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

func getElement(el interface{}, key string) (interface{}, error) {
	m, ok := el.(map[string]interface{})
	if ok {
		v, ok := m[key]
		if ok {
			return v, nil
		} else {
			return nil, fmt.Errorf("map missing key %s", key)
		}
	}
	um, ok := el.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("unable to unpack %T", el)
	}
	v, ok := um[key]
	if ok {
		return v, nil
	} else {
		return nil, fmt.Errorf("map missing key %s", key)
	}
}

func setElement(el interface{}, key string, val interface{}) error {
	m, ok := el.(map[string]interface{})
	if ok {
		m[key] = val
		return nil
	}
	um, ok := el.(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("unable to unpack %T", el)
	}
	um[key] = val
	return nil
}

func (t *Transformer) addDucatiTemplate(manifest map[string]interface{}, namePrefix string) error {
	jobsVal, ok := manifest["jobs"]
	if !ok {
		return errors.New("missing key")
	}

	jobsSlice, ok := jobsVal.([]interface{})
	if !ok {
		panic("input type not slice")
	}

	for _, jobVal := range jobsSlice {
		templates, err := getElement(jobVal, "templates")
		if err != nil {
			panic(err)
		}
		nameVal, err := getElement(jobVal, "name")
		if err != nil {
			panic(err)
		}
		name, ok := nameVal.(string)
		if !ok {
			panic("name not a string")
		}
		if !strings.HasPrefix(name, namePrefix) {
			continue
		}
		appended, err := t.AppendToSlice(templates,
			map[string]interface{}{"name": "ducati", "release": "ducati"},
		)
		if err != nil {
			panic(err)
		}
		err = setElement(jobVal, "templates", appended)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func (t *Transformer) addDucatiDBJob(manifest map[string]interface{}) error {
	jobsVal, ok := manifest["jobs"]
	if !ok {
		return errors.New("missing key")
	}
	var err error
	manifest["jobs"], err = t.AppendToSlice(jobsVal, map[string]interface{}{
		"name":            "ducati_db",
		"instances":       1,
		"persistent_disk": t.DBPersistentDisk,
		"resource_pool":   t.DBResourcePool,
		"networks": []interface{}{
			map[string]interface{}{"name": t.DBNetwork},
		},
		"templates": []interface{}{
			map[string]interface{}{"name": "postgres", "release": "ducati"},
			map[string]interface{}{"name": "consul_agent", "release": "cf"},
		},
	})
	if err != nil {
		return fmt.Errorf("adding job: %s", err)
	}
	return nil
}

func (t *Transformer) updateReleases(manifest map[string]interface{}) error {
	releasesVal, ok := manifest["releases"]
	if !ok {
		return errors.New("missing key")
	}
	var err error
	manifest["releases"], err = t.AppendToSlice(releasesVal, map[string]interface{}{
		"name":    "ducati",
		"version": t.ReleaseVersion,
	})
	if err != nil {
		return fmt.Errorf("adding release: %s", err)
	}

	return nil
}

func appendToSlice(toModify interface{}, toAppend interface{}) ([]interface{}, error) {
	asSlice, ok := toModify.([]interface{})
	if !ok {
		panic("input type not slice")
	}
	asSlice = append(asSlice, toAppend)
	return asSlice, nil
}
