package ducatify

import "fmt"

func getElement(el interface{}, key string) (interface{}, error) {
	if m, ok := el.(map[string]interface{}); ok {
		if v, ok := m[key]; ok {
			return v, nil
		}
		return nil, fmt.Errorf("map missing key %s", key)
	}

	if um, ok := el.(map[interface{}]interface{}); ok {
		if v, ok := um[key]; ok {
			return v, nil
		}
		return nil, fmt.Errorf("map missing key %s", key)
	}

	return nil, fmt.Errorf("unable to unpack %T", el)
}

func setElement(el interface{}, key string, val interface{}) error {
	if m, ok := el.(map[string]interface{}); ok {
		m[key] = val
		return nil
	}
	if um, ok := el.(map[interface{}]interface{}); ok {
		um[key] = val
		return nil
	}
	return fmt.Errorf("unable to unpack %T", el)
}

func appendToSlice(toModify interface{}, toAppend interface{}) ([]interface{}, error) {
	asSlice, ok := toModify.([]interface{})
	if !ok {
		panic("input type not slice")
	}
	asSlice = append(asSlice, toAppend)
	return asSlice, nil
}
