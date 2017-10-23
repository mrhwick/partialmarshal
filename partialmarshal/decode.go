package partialmarshal

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Unmarshal parses the JSON-encoded data and stores the result in the
// value pointed to by v.
//
// This implementation of unmarshal also detects the existence of the
// Extra struct as an embedded type in v and places any unmatching data into
// the embedded Extra map.
//
func Unmarshal(data []byte, v interface{}) error {

	// 1. Unmarshal / Decode JSON strings using the stdlib decoder.

	err := json.Unmarshal(data, v)
	if err != nil {
		return err
	}

	// 2. Identify whether the destination struct
	// contains an "Extra" substruct

	err = checkHasExtra(v)
	if err != nil {
		return err
	}

	// 3. Filter the JSON payload for fields which do not match
	// the fields of the destination struct.
	// (requires use of the reflect package)

	extraPayload, err := getExtraPayload(data, v)
	if err != nil {
		// This should never execute, but for the sake of
		// completeness, getExtraPayload returns an error.
		return err
	}

	// 4. Set the extra payload map to be the value of the
	// Extra field in the struct.

	extraField := reflect.Indirect(reflect.ValueOf(v)).FieldByName("Extra")
	extraField.Set(reflect.ValueOf(extraPayload))

	return nil
}

// getExtraPayload searches the provided JSON data for keys that do not match
// fields of the provided valud v.
// Any unmatching keys are put into a map with their respective values for return.
func getExtraPayload(data []byte, v interface{}) (map[string]interface{}, error) {
	var resultMap map[string]interface{}
	err := json.Unmarshal(data, &resultMap)
	if err != nil {
		return resultMap, err
	}

	for key := range resultMap {
		if hasFieldInStruct(v, key) {
			delete(resultMap, key)
		}
	}

	return resultMap, nil
}

// hasFieldInStruct returns true if a field matching a given key exists in the value v
// and false if there is no field matching the provided key.
func hasFieldInStruct(v interface{}, fieldKey string) bool {
	return checkHasFieldInStruct(v, fieldKey) == nil
}

// checkHasFieldInStruct returns an error if the value v is not a struct/struct pointer,
// or if there is no field in the value v that matches the provided key.
func checkHasFieldInStruct(v interface{}, fieldKey string) error {

	value := reflect.Indirect(reflect.ValueOf(v))

	if value.Kind() != reflect.Struct {
		return errors.New("value must be of type struct")
	}

	for i := 0; i < value.Type().NumField(); i++ {
		field := value.Type().Field(i)

		if strings.ToLower(field.Name) == fieldKey {
			return nil
		}

		tags := strings.Split(field.Tag.Get("json"), ",")
		for _, tag := range tags {
			if tag == fieldKey {
				return nil
			}
		}
	}

	return fmt.Errorf("could not find field %s in struct", fieldKey)
}

// checkHasExtra searches the value v for the partialmarshal.Extra type
// as a nested type. Returns an error if it does not exist on the value v, or if v
// is not a struct/struct pointer.
func checkHasExtra(v interface{}) error {

	value := reflect.Indirect(reflect.ValueOf(v))

	if value.Kind() != reflect.Struct {
		return errors.New("value must be of type struct")
	}

	extraField := value.FieldByName("Extra")
	if extraField.IsValid() && extraField.Type().String() == "partialmarshal.Extra" {
		return nil
	}

	// No matching Extra field found.
	return errors.New("no partialmarshal.Extra embedded type found in provided struct")
}
