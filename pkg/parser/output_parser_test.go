package parser

import (
	"testing"

	"github.com/archstrap/cache-server/pkg/model"
	"github.com/magiconair/properties/assert"
)

func TestParseOutputArray(t *testing.T) {
	input := &model.RespValue{
		DataType: model.TypeArray,
		Value:    []string{"ECHO", "hello", "world"},
	}

	output := ParseOutput(input)
	assert.Equal(t, output, "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n")
}
