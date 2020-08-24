package jsonutil

import (
	"testing"

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
}
`
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
}
`
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
