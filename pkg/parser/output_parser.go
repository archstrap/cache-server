package parser

import (
	"fmt"
	"strings"

	"github.com/archstrap/cache-server/pkg/model"
)

func ParseOutput(respValue *model.RespValue) string {

	switch respValue.DataType {
	case model.TypeArray:
		data, _ := respValue.Value.([]string)
		return parseArrayOutput(data)

	}

	return "+OK\r\n"

}

func ParseOutputV2(result *model.RespOutput) string {
	switch result.RespType {
	case model.TypeSimpleString, model.TypeError:
		return parseSimpleStringOutput(result)
	case model.TypeBulkString:
		return parseBulkStringOutput(result)

	}

	return "+OK\r\n"
}

func parseArrayOutput(data []string) string {

	var resultBuilder strings.Builder
	length := len(data)

	if data[0] == "COMMAND" {
		return "+OK\r\n"
	}

	resultBuilder.WriteString(fmt.Sprintf("%s%d\r\n", string(model.TypeArray), length-1))

	for i := 1; i < length; i++ {

		element := data[i]
		child := fmt.Sprintf("%s%d\r\n%s\r\n", string(model.TypeBulkString), len(element), element)
		resultBuilder.WriteString(child)

	}

	return resultBuilder.String()
}

func parseSimpleStringOutput(data *model.RespOutput) string {
	return fmt.Sprintf("%s%s\r\n", string(data.RespType), data.Data.(string))
}

func parseBulkStringOutput(data *model.RespOutput) string {
	result := data.Data.(string)
	length := len(result)
	return fmt.Sprintf("%s%d\r\n%s\r\n", string(data.RespType), length, result)
}
