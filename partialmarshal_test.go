package partialmarshal

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetReflectedValue(t *testing.T) {
	testCases := []struct {
		testDescription string
		inInterface     interface{}
		outErrMsg       string
	}{
		// Happy Path
		{
			"should return reflect.Value as expected",
			&struct {
				FieldOne string
			}{
				"some value",
			},
			"",
		},
		// Sad Path
		{
			"should return error when not pointer",
			struct {
				FieldOne string
			}{
				"some value",
			},
			"json: Unmarshal(non-pointer struct { FieldOne string })",
		},
		{
			"should return error when nil",
			nil,
			"json: Unmarshal(nil)",
		},
		{
			"should return error when pointer to non-struct kind",
			&map[string]string{"foo": "bar"},
			"json: Unmarshal(nil *map[string]string)",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			valueResult, err := getReflectedValue(tc.inInterface)
			if tc.outErrMsg != "" {
				assert.EqualError(t, err, tc.outErrMsg)
			} else {
				assert.NoError(t, err)
				expectedValud := reflect.Indirect(reflect.ValueOf(tc.inInterface))
				assert.Equal(t, expectedValud, valueResult)
			}
		})
	}
}
