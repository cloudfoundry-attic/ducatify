package ducatify

import (
	"errors"
	"fmt"
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
