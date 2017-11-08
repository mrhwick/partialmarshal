package partialmarshal

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/fatih/structs"
)

// Marshal returns the JSON encoding of v.
//
// This inmplementation of Marshal also detects the existence of
// the partialmarshal.Extra type as an embedded type in v and
// places the extra payload into the JSON output as top-level key/value pairs.
func Marshal(v interface{}) ([]byte, error) {
	// 1. Detect and retrieve the partialmarshal.Extra embedded type
	reflectedValue := reflect.Indirect(reflect.ValueOf(v))
	if reflectedValue.Kind() != reflect.Struct {
		return json.Marshal(v)
	}

	extraField := reflectedValue.FieldByName("Extra")
	if !extraField.IsValid() {
		return json.Marshal(v)
	}

	// 2. Handle any substructs that may or may not have partialmarshal.Extra fields present
	substructsMap, substructsTagMap := getSubstructsWithExtra(reflectedValue)

	// 3. Convert the value v into a map[string]interface{}
	// https://github.com/fatih/structs/issues/25
	structs.DefaultTagName = "json"
	valueAsMap := structs.Map(v)
	delete(valueAsMap, "Extra")

	// 4. Add any found substructs into the map
	for key, value := range substructsMap {
		jsonTag, found := substructsTagMap[key]
		if found {
			valueAsMap[jsonTag] = value
			delete(valueAsMap, key)
		} else {
			valueAsMap[key] = value
		}
	}

	// 5. Combine the map[string]interface{} v clone with the extra map
	extraFieldAsMap := extraField.Interface().(Extra)
	for key, value := range extraFieldAsMap {
		valueAsMap[key] = value
	}

	// 6. Encode the combined map into a JSON output
	return json.Marshal(valueAsMap)
}

func getSubstructsWithExtra(reflectedValue reflect.Value) (map[string]json.RawMessage, map[string]string) {

	substructsMap := map[string]json.RawMessage{}
	substructsTagMap := map[string]string{}

	for i := 0; i < reflectedValue.Type().NumField(); i++ {
		structField := reflectedValue.Type().Field(i)
		field := reflectedValue.Field(i)
		if field.Type().Kind() == reflect.Struct {
			encodedStruct, _ := Marshal(field.Interface())
			jsonTag, _ := parseTag(string(structField.Tag.Get("json")))
			if jsonTag != "" {
				substructsTagMap[structField.Name] = jsonTag
			}
			substructsMap[structField.Name] = encodedStruct
		}
	}
	return substructsMap, substructsTagMap
}

func parseTag(tag string) (string, string) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tag[idx+1:]
	}
	return tag, ""
}
