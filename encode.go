package partialmarshal

import (
	"encoding/json"
	"reflect"

	"github.com/fatih/structs"
)

// Marshal returns the JSON encoding of v.
//
// This inmplementation of Marshal also detects the existence of
// the partialmarshal.Extra type as an embedded type in v and
// places the extra payload into the JSON output as top-level key/value pairs.
func Marshal(v interface{}) ([]byte, error) {
	// 1. Detect and retrieve the partialmarshal.Extra embedded type

	err := checkHasExtra(v)
	if err != nil {
		return nil, err
	}

	extraField := reflect.Indirect(reflect.ValueOf(v)).FieldByName("Extra")

	// 2. Convert the value v into a map[string]interface{}

	// https://github.com/fatih/structs/issues/25
	structs.DefaultTagName = "json"
	valueAsMap := structs.Map(v)
	delete(valueAsMap, "Extra")

	// 3. Combine the map[string]interface{} v clone with the extra map

	extraFieldAsMap := extraField.Interface().(Extra)
	for key, value := range extraFieldAsMap {
		valueAsMap[key] = value
	}

	// 4. Encode the combined map into a JSON output
	return json.Marshal(valueAsMap)
}
