package updater

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestYAMLToJSON(t *testing.T) {
	yamlText := `
no: no
"yes": "yes"
true_bool: true
true_text: "true"
true: key
null_null: ~
null_text: "~"
sub:
  no: no
  "yes": "yes"
  true_bool: true
  true_text: "true"
  true: key
  null_null: ~
  null_text: "~"
list:
  - one
  - "two"
  - true
  - "true"
  - yes
  - ~
  - "~"
`
	jsonText := `
{
	"no": "no",
	"yes": "yes",
	"true_bool": true,
	"true_text": "true",
	"true": "key",
	"null_null": null,
	"null_text": "~",
	"sub": {
		"no": "no",
		"yes": "yes",
		"true_bool": true,
		"true_text": "true",
		"true": "key",
		"null_null": null,
		"null_text": "~"
	},
	"list": ["one", "two", true, "true", "yes", null, "~"]
}
`
	j, err := YAMLToJSON([]byte(yamlText))
	if err != nil {
		t.Fatal(err)
	}
	var yamlData interface{}
	err = json.Unmarshal(j, &yamlData)
	if err != nil {
		t.Fatal(err)
	}
	var jsonData interface{}
	err = json.Unmarshal([]byte(jsonText), &jsonData)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(j))
	if !reflect.DeepEqual(yamlData, jsonData) {
		t.Errorf("yaml and json not equal, yaml in json: %s", string(j))
	}
}

func TestFloat(t *testing.T) {
	j, err := YAMLToJSON([]byte(`[0.0, .inf, 3.141592653589793238462643383279, 685_230.15]`))
	if err != nil {
		t.Fatal(err)
	}
	if string(j) != `[0.0,null,3.141592653589793238462643383279,685230.15]` {
		t.Errorf("float failed: %s", string(j))
	}
}
