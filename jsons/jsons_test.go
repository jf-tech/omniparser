package jsons

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"
)

func TestPrettyMarshal_Success(t *testing.T) {
	actual, err := PrettyMarshal(struct {
		Name    string
		Age     int
		Smoker  bool
		Hobbies []string
	}{
		Name:    "John",
		Age:     65,
		Smoker:  false,
		Hobbies: []string{"flying", "fishing", "reading"},
	})
	assert.NoError(t, err)
	expected := `{
	"Name": "John",
	"Age": 65,
	"Smoker": false,
	"Hobbies": [
		"flying",
		"fishing",
		"reading"
	]
}`
	assert.Equal(t, expected, actual)
}

func TestPrettyMarshal_Failure(t *testing.T) {
	_, err := PrettyMarshal(struct {
		Name           string
		Unmarshallable func() bool
	}{
		Name:           "John",
		Unmarshallable: func() bool { return true },
	})
	assert.Error(t, err)
}

func TestBestEffortPrettyMarshal_Success(t *testing.T) {
	actual := BestEffortPrettyMarshal(struct {
		Name    string
		Age     int
		Smoker  bool
		Hobbies []string
	}{
		Name:    "John",
		Age:     65,
		Smoker:  false,
		Hobbies: []string{"flying", "fishing", "reading"},
	})
	expected := `{
	"Name": "John",
	"Age": 65,
	"Smoker": false,
	"Hobbies": [
		"flying",
		"fishing",
		"reading"
	]
}`
	assert.Equal(t, expected, actual)
}

func TestBestEffortPrettyMarshal_Failure(t *testing.T) {
	assert.Equal(t, "{}", BestEffortPrettyMarshal(
		struct {
			Name           string
			Unmarshallable func() bool
		}{
			Name:           "John",
			Unmarshallable: func() bool { return true },
		}))
}

func TestBPMInSnapshot(t *testing.T) {
	cupaloy.SnapshotT(t, BPM(struct {
		Name   string
		IntV   int
		StrV   string
		SliceV []float64
		MapV   map[string]interface{}
	}{
		Name:   "a snapshot",
		IntV:   123,
		StrV:   `a string with "quotes"`,
		SliceV: []float64{3.14159, 2.71828},
		MapV: map[string]interface{}{
			"pi":    3.14159,
			"linux": "Linux Is Not UniX",
		},
	}))
}

func TestPrettyJSON_Success(t *testing.T) {
	pretty, err := PrettyJSON(`{"Name":"John","Age":65,"Smoker":false,"Hobbies":["flying","fishing","reading"]}`)
	assert.NoError(t, err)
	cupaloy.SnapshotT(t, pretty)
}

func TestPrettyJSON_Failure(t *testing.T) {
	pretty, err := PrettyJSON(`{`)
	assert.Error(t, err)
	assert.Equal(t, "unexpected end of JSON input", err.Error())
	assert.Equal(t, "", pretty)
}

func TestBestEffortPrettyJSON_Success(t *testing.T) {
	pretty := BestEffortPrettyJSON(`{"Name":"John","Age":65,"Smoker":false,"Hobbies":["flying","fishing","reading"]}`)
	cupaloy.SnapshotT(t, pretty)
}

func TestBestEffortPrettyJSON_Failure(t *testing.T) {
	pretty := BestEffortPrettyJSON(`{`)
	assert.Equal(t, "{}", pretty)
}

func TestBPJ_Success(t *testing.T) {
	pretty := BPJ(`{"Name":"John","Age":65,"Smoker":false,"Hobbies":["flying","fishing","reading"]}`)
	cupaloy.SnapshotT(t, pretty)
}
