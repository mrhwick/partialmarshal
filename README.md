# PartialMarshal
A Go library for JSON marshaling  with extra payloads.

Standard Go JSON marshaling into a struct discards any unmatching fields from the JSON payload. In some use cases, the developer does not want to discard this extra payload data. This library provides a familiar interface for performing marshaling and unmarshaling while keeping any extra data for use by the calling code.

## Install

```bash
go get github.com/skuid/partialmarshal
```

## Usage and Examples

Just like the standard library `json` package, `partialmarshal` provides a simple pair of functions for marshaling and unmarshaling of JSON-formatted data into and out of structs.

When a user wants to use partialmarshal to hold onto extra data from a JSON payload, they may simply add the `partialmarshal.Extra` type as an embedded type in their struct. This embedded type is used by the partialmarshal library for storage of any data from the provided JSON that doesn't match a field of that struct.

```go
type Person struct {
	Name string
	FavoriteFood string `json:"favorite_food"`
	partialmarshal.Extra
}

jsonData := []bytes(`{"name": "gopher", "favorite_food": "Pizza", "age": 25}`)
```

```go
// Unmarshaling
// => Person{Name: "gopher", FavoriteFood: "Pizza", partialmarshal.Extra{"age":25}}
var p Person
partialmarshal.Unmarshal(jsonData, &p)

// Marshaling
// => `{"name": "gopher", "favorite_food": "Salad", "age" 25}`
p.FavoriteFood = "Salad"
result, err := partialmarshal.Marshal(p)
```
