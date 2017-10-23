package partialmarshal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshal(t *testing.T) {
	testCases := []struct {
		testDescription string
		inData          []byte
		inStruct        interface{}
		outStruct       interface{}
		outErrMsg       string
	}{
		// Happy Path Cases
		{
			"should unmarshal extra payload into extra field",
			[]byte(`{"field_one": "value one", "field_two": "value two"}`),
			&struct {
				FieldOne string `json:"field_one"`
				Extra
			}{},
			&struct {
				FieldOne string `json:"field_one"`
				Extra
			}{
				"value one",
				map[string]interface{}{
					"field_two": "value two",
				},
			},
			"",
		},
		// Sad Path Cases
		{
			"should return error when provided value not struct pointer",
			[]byte(`{"field_one": "value one", "field_two": "value two"}`),
			"",
			&struct{}{},
			"json: Unmarshal(non-pointer string)",
		},
		{
			"should return error when provided with malformed JSON",
			[]byte(`decidedly not json in format`),
			&struct{}{},
			&struct{}{},
			"invalid character 'd' looking for beginning of value",
		},
		{
			"should return error when partialmarshal.Extra not nested in provided struct",
			[]byte(`{"field_one": "value one", "field_two": "value two"}`),
			&struct{}{},
			&struct{}{},
			"no partialmarshal.Extra embedded type found in provided struct",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {

			err := Unmarshal(tc.inData, tc.inStruct)
			if tc.outErrMsg != "" {
				assert.EqualError(t, err, tc.outErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.inStruct, tc.outStruct)
			}
		})
	}
}

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

func TestCheckHasFieldInStruct(t *testing.T) {
	testCases := []struct {
		testDescription string
		inStruct        interface{}
		inKey           string
		outErrMsg       string
	}{
		// Happy Path Cases
		{
			"Should detect field when matching field name & no JSON tag",
			struct {
				TestFieldOne string
			}{},
			"testfieldone",
			"",
		},
		{
			"Should detect field when matching field name & no matching JSON tag",
			struct {
				TestFieldtwo string `json:"some_test_field_two"`
			}{},
			"testfieldtwo",
			"",
		},
		{
			"Should detect field when no matching field name & matching JSON tag",
			struct {
				TestFieldThree string `json:"test_field_three"`
			}{},
			"test_field_three",
			"",
		},
		{
			"Should detect field when matching field name & matching JSON tag",
			struct {
				TestFieldFour string `json:"testfieldfour"`
			}{},
			"testfieldfour",
			"",
		},
		{
			"Should detect field when matching field name & matching JSON tag w/ multiple JSON tags",
			struct {
				TestFieldFour string `json:"testfieldfour,omitempty"`
			}{},
			"testfieldfour",
			"",
		},
		{
			"Should detect field when matching field name & no JSON tag w/ struct pointer value",
			&struct {
				TestFieldFive string
			}{},
			"testfieldfive",
			"",
		},
		// Sad Path Cases
		{
			"Should not detect field when no matching field name & no matching JSON tag",
			struct {
				TestfieldSix string `json:"test_field_six"`
			}{},
			"some_unmatching_field_name",
			"could not find field some_unmatching_field_name in struct",
		},
		{
			"Should not detect field when struct has no fields (doesn't make sense, but whatever)",
			struct{}{},
			"some_unmatching_field_name",
			"could not find field some_unmatching_field_name in struct",
		},
		{
			"Should throw error when provided interface not a struct",
			8675309,
			"does_not_matter_what_field_we_look_for_anymore",
			"value must be of type struct",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			err := checkHasFieldInStruct(tc.inStruct, tc.inKey)
			if tc.outErrMsg != "" {
				assert.EqualError(t, err, tc.outErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetExtraPayload(t *testing.T) {

	testCases := []struct {
		testDescription string
		inData          []byte
		inStruct        interface{}
		outPayload      map[string]interface{}
		outErrMsg       string
	}{
		// Happy Path Cases
		{
			"Should return map with extra payload for JSON object with unmatching fields",
			[]byte(`{"some_field": "some value"}`),
			struct{}{},
			map[string]interface{}{
				"some_field": "some value",
			},
			"",
		},
		{
			"Should return empty map for JSON object with no extra payload",
			[]byte(`{"some_field": "some value", "somesecondfield": "some second value"}`),
			struct {
				SomeField       string `json:"some_field"`
				SomeSecondField string
			}{},
			map[string]interface{}{},
			"",
		},
		{
			"Should return empty map for empty JSON object (implicit no extra payload)",
			[]byte(`{}`),
			struct {
				thing string
			}{},
			map[string]interface{}{},
			"",
		},
		// Sad Path Cases
		{
			"Should return error for malformed JSON data",
			[]byte(`json is easy to malform`),
			struct {
				SomeField string `json:"some_field"`
			}{},
			map[string]interface{}{},
			"invalid character 'j' looking for beginning of value",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {

			result, err := getExtraPayload(tc.inData, tc.inStruct)

			if tc.outErrMsg != "" {
				assert.EqualError(t, err, tc.outErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.outPayload, result)
			}
		})
	}
}
