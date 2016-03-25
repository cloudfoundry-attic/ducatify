package ducatify

type Transformer struct{}

func (t *Transformer) Transform(manifest map[string]interface{}) error {
	releasesVal, ok := manifest["releases"]
	if !ok {
		panic("missing releases")
	}
	releases, ok := releasesVal.([]interface{})
	if !ok {
		panic("bad type for releases")
	}
	releases = append(releases, map[string]interface{}{
		"name":    "ducati",
		"version": "latest",
	})
	manifest["releases"] = releases

	return nil
}
