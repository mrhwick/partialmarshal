package partialmarshal

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
)

// Unmarshal parses the JSON-encoded data and stores the result in the
// value pointed to by v.
//
// This implementation of Unmarshal also detects the existence of the
// partialmarshal.Extra type as an embedded type in v and places any
// unmatching data into the embedded Extra map.
//
func Unmarshal(data []byte, v interface{}) error {
	if bytes.HasPrefix(data, []byte("[")) {
		return unmarshalArray(data, v)
	}
	if bytes.HasPrefix(data, []byte("{")) {
		return unmarshalObject(data, v)
	}
	return json.Unmarshal(data, &v)
}

func unmarshalArray(data json.RawMessage, v interface{}) error {
	var JSONObjectList []json.RawMessage
	json.Unmarshal(data, &JSONObjectList)
	reflectedValue := reflect.ValueOf(v)
	if reflectedValue.Kind() != reflect.Ptr || reflectedValue.IsNil() {
		return &json.InvalidUnmarshalError{
			Type: reflect.TypeOf(v),
		}
	}
	reflectedValue = reflect.Indirect(reflectedValue)
	if reflectedValue.Kind() != reflect.Slice {
		return &json.InvalidUnmarshalError{
			Type: reflect.TypeOf(v),
		}
	}
	for _, obj := range JSONObjectList {
		sliceElementType := reflectedValue.Type().Elem()
		temporarySliceElement := reflect.New(sliceElementType).Interface()
		err := Unmarshal(obj, temporarySliceElement)
		if err != nil {
			return err
		}
		reflectedValue.Set(reflect.Append(reflectedValue, reflect.Indirect(reflect.ValueOf(temporarySliceElement))))
	}
	return nil
}

func unmarshalObject(data json.RawMessage, v interface{}) error {
	// 1. Check for a valid pointer to value of kind struct.
	reflectedValue, err := getReflectedValue(v)
	if err != nil {
		return err
	}

	// 2. Create the json.RawMessage map of this JSON object
	var rawMap map[string]json.RawMessage
	err = json.Unmarshal(data, &rawMap)
	if err != nil {
		return err
	}

	// 3. Decode matching data into the struct and recursively call for substructs
	err = decodeMatching(rawMap, reflectedValue)
	if err != nil {
		return err
	}

	// 4. Put Extra values into the Extra nested struct
	extraField := reflectedValue.FieldByName("Extra")
	if extraField.IsValid() {
		extraField.Set(reflect.ValueOf(rawMap))
	}

	return nil
}

func popValueByField(rawMap map[string]json.RawMessage, field reflect.StructField) (json.RawMessage, bool) {
	rawValue, found := rawMap[field.Name]
	if !found {
		// Attempt match by JSON tags.
		tags := strings.Split(field.Tag.Get("json"), ",")
		for _, tag := range tags {
			rawValue, found = rawMap[tag]
			if found {
				delete(rawMap, tag)
				break
			}
		}

		if !found {
			// Still no match found, continue to next struct field
			return rawValue, false
		}
	} else {
		delete(rawMap, field.Name)
	}
	return rawValue, true
}

func decodeMatching(rawMap map[string]json.RawMessage, reflectedValue reflect.Value) error {
	for i := 0; i < reflectedValue.Type().NumField(); i++ {
		field := reflectedValue.Type().Field(i)
		// Attempt match by field.Name
		rawValue, found := popValueByField(rawMap, field)
		if !found {
			continue
		}

		temp := reflect.New(field.Type).Interface()

		if field.Type.Kind() == reflect.Struct || field.Type.Kind() == reflect.Slice {
			err := Unmarshal(rawValue, temp)
			if err != nil {
				return err
			}
		} else {
			err := json.Unmarshal(rawValue, &temp)
			if err != nil {
				return err
			}
		}

		actualValue := reflect.Indirect(reflect.ValueOf(temp))
		reflectedValue.FieldByName(field.Name).Set(actualValue)

	}
	return nil
}
