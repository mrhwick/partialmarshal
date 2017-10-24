package partialmarshal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckHasExtra(t *testing.T) {

	testCases := []struct {
		testDescription string
		in              interface{}
		outErrMsg       string
	}{
		// Happy Path Cases
		{
			"Should return nil for struct type with Extra substruct",
			struct {
				someOtherField string
				Extra
			}{},
			"",
		},
		{
			"Should return nil for struct pointer type with Extra substruct",
			&struct {
				someOtherField string
				Extra
			}{},
			"",
		},
		// Sad Path Cases
		{
			"Should return error for struct type without Extra substruct",
			struct {
				someOtherField string
			}{},
			"no partialmarshal.Extra embedded type found in provided struct",
		},
		{
			"Should return error for struct pointer type without Extra substruct",
			&struct {
				someOtherField string
			}{},
			"no partialmarshal.Extra embedded type found in provided struct",
		},
		{
			"Should return error for non-struct type string",
			"",
			"value must be of type struct",
		},
		{
			"Should return error for non-struct type number",
			1990,
			"value must be of type struct",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			err := checkHasExtra(tc.in)

			if tc.outErrMsg != "" {
				assert.EqualError(t, err, tc.outErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
