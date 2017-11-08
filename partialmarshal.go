package partialmarshal

import (
	"encoding/json"
	"reflect"
)

// Extra - A type provided for use as an embedded type to indicate
// a storage location for extra payloads when unmarshaling.
type Extra map[string]json.RawMessage

func getReflectedValue(v interface{}) (reflect.Value, error) {
	reflectedValue := reflect.ValueOf(v)
	if reflectedValue.Kind() != reflect.Ptr || reflectedValue.IsNil() {
		// Invalid because either Nil or Non-Pointer
		return reflectedValue, &json.InvalidUnmarshalError{
			Type: reflect.TypeOf(v),
		}
	}
	reflectedValue = reflect.Indirect(reflectedValue)
	if reflectedValue.Kind() != reflect.Struct {
		// Invalid because not a struct
		return reflectedValue, &json.InvalidUnmarshalError{
			Type: reflect.TypeOf(v),
		}
	}
	return reflectedValue, nil
}
