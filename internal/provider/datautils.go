package provider

import (
	"encoding/json"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func setValues(d *schema.ResourceData, values map[string]interface{}) error {
	for k, v := range values {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

// emptyIfNil returns an empty non-nil slice when s is nil; otherwise returns s.
// The v4 SDK getters return nil for unset slices, but terraform-plugin-sdk treats
// nil and empty list differently in state, causing spurious plan diffs.
func emptyIfNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

// stringOrDefault returns def when s is empty; otherwise s. Used to surface
// Algolia server-side defaults in state (the v4 API omits fields at default values).
func stringOrDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func castStringList(list interface{}) []string {
	// we are initializing non nil array to be marshaled to [] in JSON
	strs := []string{}
	for _, v := range list.([]interface{}) {
		strs = append(strs, v.(string))
	}
	return strs
}

func castStringSet(set interface{}) []string {
	// we are initializing non nil array to be marshaled to [] in JSON
	strs := []string{}
	for _, v := range set.(*schema.Set).List() {
		strs = append(strs, v.(string))
	}
	return strs
}

func castStringMap(m interface{}) map[string]string {
	strMap := map[string]string{}
	for k, v := range m.(map[string]interface{}) {
		strMap[k] = v.(string)
	}
	return strMap
}

func castInterfaceMap(m interface{}) map[string]interface{} {
	interfaceMap := map[string]interface{}{}
	for k, v := range m.(map[string]interface{}) {
		interfaceMap[k] = v
	}
	return interfaceMap
}

func diffJsonSuppress(k, old, new string, d *schema.ResourceData) bool {
	result, _ := jsonBytesEqual([]byte(old), []byte(new))
	return result
}

// jsonBytesEqual compares the JSON in two byte slices
func jsonBytesEqual(a, b []byte) (bool, error) {
	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false, err
	}
	return mapsEqual(j, j2), nil
}

func mapsEqual(m1, m2 interface{}) bool {
	return reflect.DeepEqual(m2, m1)
}
