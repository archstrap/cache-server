package util

import (
	"testing"

	"github.com/archstrap/cache-server/pkg/model"
)

func TestGetBytes(t *testing.T) {
	testCases := []struct {
		name     string
		input    *model.RespValue
		expected int
	}{
		{
			name: "PING command",
			input: &model.RespValue{
				DataType: model.TypeArray,
				Value:    []string{"PING"},
			},
			expected: 14,
		},
		{
			name: "SET command",
			input: &model.RespValue{
				DataType: model.TypeArray,
				Value:    []string{"SET", "foo", "1"},
			},
			expected: 29,
		},
		{
			name: "REPLCONF command",
			input: &model.RespValue{
				DataType: model.TypeArray,
				Value:    []string{"REPLCONF", "GETACK", "*"},
			},
			expected: 37,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			actual := GetBytes(tt.input)
			if actual != tt.expected {
				t.Errorf("Name: %s \n Actual:%d, but Expected:%d", tt.name, actual, tt.expected)
			}
		})
	}
}
