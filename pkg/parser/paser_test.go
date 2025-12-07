package parser

import (
	"strings"
	"testing"

	"github.com/archstrap/cache-server/pkg/model"
	"github.com/magiconair/properties/assert"
)

func TestParseInteger(t *testing.T) {

	reader := strings.NewReader(":2\r\n")

	value, err := Parse(reader)

	assert.Equal(t, err, nil)
	assert.Equal(t, value.Value, 2)
	assert.Equal(t, value.DataType, model.TypeInteger)

}

func TestParseSimpleString(t *testing.T) {

	reader := strings.NewReader("+OK\r\n")

	value, err := Parse(reader)

	assert.Equal(t, err, nil)
	assert.Equal(t, value.Value, "OK")
	assert.Equal(t, value.DataType, model.TypeSimpleString)

}

func TestParseBulkString(t *testing.T) {
	reader := strings.NewReader("$3\r\nGET\r\n")

	value, err := Parse(reader)
	assert.Equal(t, err, nil)
	assert.Equal(t, value.Value, "GET")
	assert.Equal(t, value.DataType, model.TypeBulkString)

}

func TestParseArray(t *testing.T) {
	reader := strings.NewReader("*1\r\n$3\r\nGET\r\n")

	value, err := Parse(reader)

	assert.Equal(t, err, nil)
	// TODO add matches
	assert.Equal(t, value.DataType, model.TypeArray)

}
