package jsonutil

import (
	"bytes"
	"encoding/json"
)

const (
	prettyIndent = "\t"
)

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

func BestEffortPrettyMarshal(v interface{}) string {
	jsonStr, err := PrettyMarshal(v)
	if err != nil {
		return "{}"
	}
	return jsonStr
}
