package partialmarshal

import (
	"errors"
	"reflect"
)

// Extra - A type provided for use as an embedded type to indicate
// a storage location for extra payloads when unmarshaling.
type Extra map[string]interface{}

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
