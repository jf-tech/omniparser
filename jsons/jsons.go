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

// BPM is a shortcut (mostly used ing tests) to BestEffortPrettyMarshal.
var BPM = BestEffortPrettyMarshal
