package partialmarshal

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleUnmarshal() {
	// A JSON-formatted string
	JSONData := []byte(`{
		"ExampleFieldOne": "value 1",
		"example_field_two": "value 2",
		"some_other_field": "some other value"
	}`)

	// A struct type with partialmarshal.Extra included as an embedded type
	type StructWithExtra struct {
		ExampleFieldOne string
		ExampleFieldTwo string `json:"example_field_two"`
		Extra
	}

	// A struct type without partialmarshal.Extra included as an embedded type
	type StructWithoutExtra struct {
		ExampleFieldOne string
		ExampleFieldTwo string `json:"example_field_two"`
	}

	fmt.Println("Nominal Case:")
	var destination StructWithExtra
	err := Unmarshal(JSONData, &destination)
	fmt.Println(err)
	fmt.Println(destination.ExampleFieldOne)
	fmt.Println(destination.ExampleFieldTwo)
	fmt.Printf("%#v", destination.Extra)

	fmt.Println("\n\nError Case:")
	var badDestination StructWithoutExtra
	err = Unmarshal(JSONData, &badDestination)
	fmt.Println(err)
	// Output:
	// Nominal Case:
	// <nil>
	// value 1
	// value 2
	// partialmarshal.Extra{"some_other_field":"some other value"}
	//
	// Error Case:
	// no partialmarshal.Extra embedded type found in provided struct
}

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
		{
			"Should detect field w/ same case when matching field name & no JSON tag w/ struct pointer value",
			&struct {
				TestFieldFive string
			}{},
			"TestFieldFive",
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
