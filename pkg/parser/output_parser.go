package parser

import (
	"fmt"
	"strings"
)

func ParseOutput(respValue *RespValue) string {

	switch respValue.DataType {
	case TypeArray:
		data, _ := respValue.Value.([]string)
		return parseArrayOutput(data)

	}

	return "+OK\r\n"

}

func parseArrayOutput(data []string) string {

	var resultBuilder strings.Builder
	length := len(data)

	if data[0] == "COMMAND" {
		return "+OK\r\n"
	}

	resultBuilder.WriteString(fmt.Sprintf("%s%d\r\n", string(TypeArray), length-1))

	for i := 1; i < length; i++ {

		element := data[i]
		child := fmt.Sprintf("%s%d\r\n%s\r\n", string(TypeBulkString), len(element), element)
		resultBuilder.WriteString(child)

	}

	return resultBuilder.String()
}
