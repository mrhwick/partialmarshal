package partialmarshal

import (
	"encoding/json"
	"fmt"
	"reflect"
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
	fmt.Printf("%s", destination.Extra["some_other_field"])

	// Output:
	// Nominal Case:
	// <nil>
	// value 1
	// value 2
	// "some other value"
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
				map[string]json.RawMessage{
					"field_two": []byte("\"value two\""),
				},
			},
			"",
		},
		{
			"should unmarshal extra payload into extra field of substruct",
			[]byte(`{"field_one": "value one", "FieldSubStruct": {"sub_field_one": "sub value one", "sub_field_two": "sub value two"}, "field_two": "value two"}`),
			&struct {
				FieldOne       string `json:"field_one"`
				FieldSubStruct struct {
					SubFieldOne string `json:"sub_field_one"`
					Extra
				}
				Extra
			}{},
			&struct {
				FieldOne       string `json:"field_one"`
				FieldSubStruct struct {
					SubFieldOne string `json:"sub_field_one"`
					Extra
				}
				Extra
			}{
				"value one",
				struct {
					SubFieldOne string `json:"sub_field_one"`
					Extra
				}{
					"sub value one",
					map[string]json.RawMessage{
						"sub_field_two": []byte("\"sub value two\""),
					},
				},
				map[string]json.RawMessage{
					"field_two": []byte("\"value two\""),
				},
			},
			"",
		},
		{
			"should unmarshal extra payload into extra field of substruct with matching json tags",
			[]byte(`{"field_one": "value one", "FieldSubStruct": {"sub_field_one": "sub value one", "sub_field_two": "sub value two"}, "field_two": "value two"}`),
			&struct {
				FieldOne       string `json:"field_one"`
				FieldSubStruct struct {
					SubFieldOne string `json:"sub_field_one"`
					Extra
				}
				Extra
			}{},
			&struct {
				FieldOne       string `json:"field_one"`
				FieldSubStruct struct {
					SubFieldOne string `json:"sub_field_one"`
					Extra
				}
				Extra
			}{
				"value one",
				struct {
					SubFieldOne string `json:"sub_field_one"`
					Extra
				}{
					"sub value one",
					map[string]json.RawMessage{
						"sub_field_two": []byte("\"sub value two\""),
					},
				},
				map[string]json.RawMessage{
					"field_two": []byte("\"value two\""),
				},
			},
			"",
		},
		{
			"should unmarshal extra payload into extra field of substruct with matching substruct json tags",
			[]byte(`{"field_one": "value one", "field_sub_struct": {"sub_field_one": "sub value one", "sub_field_two": "sub value two"}, "field_two": "value two"}`),
			&struct {
				FieldOne       string `json:"field_one"`
				FieldSubStruct struct {
					SubFieldOne string `json:"sub_field_one"`
					Extra
				} `json:"field_sub_struct"`
				Extra
			}{},
			&struct {
				FieldOne       string `json:"field_one"`
				FieldSubStruct struct {
					SubFieldOne string `json:"sub_field_one"`
					Extra
				} `json:"field_sub_struct"`
				Extra
			}{
				"value one",
				struct {
					SubFieldOne string `json:"sub_field_one"`
					Extra
				}{
					"sub value one",
					map[string]json.RawMessage{
						"sub_field_two": []byte("\"sub value two\""),
					},
				},
				map[string]json.RawMessage{
					"field_two": []byte("\"value two\""),
				},
			},
			"",
		},
		{
			"should unmarshal extra payload into extra field of multiple substructs with matching substruct json tags",
			[]byte(`{"field_one": "value one", "field_sub_struct_one": {"sub_field_one": "sub value one", "sub_field_two": "sub value two"}, "field_sub_struct_two": {"sub_field_three": "sub value three", "sub_field_four": "sub value four"}, "field_two": "value two"}`),
			&struct {
				FieldOne          string `json:"field_one"`
				FieldSubStructOne struct {
					SubFieldOne string `json:"sub_field_one"`
					Extra
				} `json:"field_sub_struct_one"`
				FieldSubStructTwo struct {
					SubFieldThree string `json:"sub_field_three"`
					Extra
				} `json:"field_sub_struct_two"`
				Extra
			}{},
			&struct {
				FieldOne          string `json:"field_one"`
				FieldSubStructOne struct {
					SubFieldOne string `json:"sub_field_one"`
					Extra
				} `json:"field_sub_struct_one"`
				FieldSubStructTwo struct {
					SubFieldThree string `json:"sub_field_three"`
					Extra
				} `json:"field_sub_struct_two"`
				Extra
			}{
				"value one",
				struct {
					SubFieldOne string `json:"sub_field_one"`
					Extra
				}{
					"sub value one",
					map[string]json.RawMessage{
						"sub_field_two": []byte("\"sub value two\""),
					},
				},
				struct {
					SubFieldThree string `json:"sub_field_three"`
					Extra
				}{
					"sub value three",
					map[string]json.RawMessage{
						"sub_field_four": []byte("\"sub value four\""),
					},
				},
				map[string]json.RawMessage{
					"field_two": []byte("\"value two\""),
				},
			},
			"",
		},
		{
			"should unmarshal normally without storing extra if Extra not embedded",
			[]byte(`{"field_one": "value one", "field_two": "value two"}`),
			&struct {
				FieldOne string `json:"field_one"`
			}{},
			&struct {
				FieldOne string `json:"field_one"`
			}{
				"value one",
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

func TestPopValueByField(t *testing.T) {
	testCases := []struct {
		testDescription string
		inMap           map[string]json.RawMessage
		inField         reflect.StructField
		inIdentifier    string
		outRawValue     json.RawMessage
		outFound        bool
	}{
		// Happy Path
		{
			"should find value matching field name",
			map[string]json.RawMessage{
				"FieldOne": []byte("\"value one\""),
			},
			reflect.StructField{
				Name: "FieldOne",
			},
			"FieldOne",
			[]byte(`"value one"`),
			true,
		},
		{
			"should find value matching field tags",
			map[string]json.RawMessage{
				"field_two": []byte("\"value two\""),
			},
			reflect.StructField{
				Name: "FieldTwo",
				Tag:  `json:"field_two"`,
			},
			"field_two",
			[]byte(`"value two"`),
			true,
		},
		// Sad Path
		{
			"should return false found for no matching field",
			map[string]json.RawMessage{
				"field_three": []byte("\"value three\""),
			},
			reflect.StructField{
				Name: "FieldTwo",
				Tag:  `json:"field_two"`,
			},
			"field_three",
			[]byte(nil),
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			value, found := popValueByField(tc.inMap, tc.inField)
			assert.Equal(t, tc.outRawValue, value)
			assert.Equal(t, tc.outFound, found)
			if tc.outFound {
				// Should delete the field value from the map
				assert.Nil(t, tc.inMap[tc.inIdentifier])
			}
		})
	}
}

func TestDecodeMatching(t *testing.T) {
	type testStruct struct {
		FieldOne string `json:"field_one"`
		FieldTwo int    `json:"field_two"`
	}
	testCases := []struct {
		testDescription string
		inMap           map[string]json.RawMessage
		inStruct        testStruct
		outStruct       testStruct
		outErrMsg       string
	}{
		// Happy Path
		{
			"should fill value matching field name",
			map[string]json.RawMessage{
				"FieldOne": []byte(`"value one"`),
			},
			testStruct{},
			testStruct{
				"value one",
				0,
			},
			"",
		},
		{
			"should fill value matching json tag",
			map[string]json.RawMessage{
				"field_one": []byte(`"value one"`),
			},
			testStruct{},
			testStruct{
				"value one",
				0,
			},
			"",
		},
		{
			"should fill multiple values matching field name",
			map[string]json.RawMessage{
				"FieldOne": []byte(`"value one"`),
				"FieldTwo": []byte(`3`),
			},
			testStruct{},
			testStruct{
				"value one",
				3,
			},
			"",
		},
		{
			"should fill multiple values matching json tag",
			map[string]json.RawMessage{
				"field_one": []byte(`"value one"`),
				"FieldTwo":  []byte(`4`),
			},
			testStruct{},
			testStruct{
				"value one",
				4,
			},
			"",
		},
		// Sad Path
		{
			"should return error on bad JSON formatting",
			map[string]json.RawMessage{
				"field_one": []byte(`certainly not a raw json string`),
			},
			testStruct{},
			testStruct{},
			"invalid character 'c' looking for beginning of value",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			indirectedValue := reflect.Indirect(reflect.ValueOf(&tc.inStruct))
			err := decodeMatching(tc.inMap, indirectedValue)
			if tc.outErrMsg != "" {
				assert.EqualError(t, err, tc.outErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.outStruct, tc.inStruct)
			}
		})
	}
}
