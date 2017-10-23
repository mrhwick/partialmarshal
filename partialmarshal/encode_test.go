package partialmarshal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	testCases := []struct {
		testDescription string
		inStruct        interface{}
		outData         []byte
		outErrMsg       string
	}{
		// Happy Path Cases
		{
			"should marshal extra field into top-level keys in payload from struct pointer",
			&struct {
				FieldOne string
				Extra
			}{
				"value one",
				Extra{
					"field_two": "value two",
				},
			},
			[]byte(`{"FieldOne":"value one","field_two":"value two"}`),
			"",
		},
		{
			"should marshal extra field into top-level keys in payload from non-pointer struct",
			struct {
				FieldOne string
				Extra
			}{
				"value one",
				Extra{
					"field_two": "value two",
				},
			},
			[]byte(`{"FieldOne":"value one","field_two":"value two"}`),
			"",
		},
		{
			"should marshal struct fields using json tags",
			&struct {
				FieldOne string `json:"field_one"`
				Extra
			}{
				"value one",
				Extra{
					"field_two": "value two",
				},
			},
			[]byte(`{"field_one":"value one","field_two":"value two"}`),
			"",
		},
		// Sad Path Cases
		{
			"should return error when no partialmarshal.Extra embedded type present",
			&struct {
				FieldOne string `json:"field_one"`
			}{
				"value one",
			},
			nil,
			"no partialmarshal.Extra embedded type found in provided struct",
		},
		{
			"should return error when provided value not struct/struct pointer",
			"",
			nil,
			"value must be of type struct",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {

			result, err := Marshal(tc.inStruct)
			if tc.outErrMsg != "" {
				assert.EqualError(t, err, tc.outErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.outData, result)
			}
		})
	}
}
