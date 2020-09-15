package jsons

import (
	"bytes"
	"encoding/json"
)

const (
	prettyIndent = "\t"
)

// PrettyMarshal does a JSON marshaling of 'v' with human readable output.
func PrettyMarshal(v interface{}) (string, error) {
	valueBuf := new(bytes.Buffer)
	enc := json.NewEncoder(valueBuf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", prettyIndent)
	err := enc.Encode(v)
	if err != nil {
		return "", err
	}
	return valueBuf.String(), nil
}

// BestEffortPrettyMarshal does a best effort JSON marshaling of 'v' with human
// readable output. '{}' will be produced if there is any JSON marshal error. This
// function never fails.
func BestEffortPrettyMarshal(v interface{}) string {
	jsonStr, err := PrettyMarshal(v)
	if err != nil {
		return "{}"
	}
	return jsonStr
}

// BPM is a shortcut (mostly used in tests) to BestEffortPrettyMarshal.
var BPM = BestEffortPrettyMarshal

// PrettyJSON reformats a json string to be pretty
func PrettyJSON(jsonStr string) (string, error) {
	var v interface{}
	err := json.Unmarshal([]byte(jsonStr), &v)
	if err != nil {
		return "", err
	}
	return PrettyMarshal(v)
}

// BestEffortPrettyJSON reformats a json string to be pretty, ignoring any error.
func BestEffortPrettyJSON(jsonStr string) string {
	s, err := PrettyJSON(jsonStr)
	if err != nil {
		return "{}"
	}
	return s
}

// BPJ is a shortcut (mostly used in tests) to BestEffortPrettyJSON.
var BPJ = BestEffortPrettyJSON
