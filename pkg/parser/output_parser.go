package parser

import (
	"fmt"
	"strings"

	"github.com/archstrap/cache-server/pkg/model"
)

func ParseOutput(result *model.RespOutput) string {
	switch result.RespType {
	case model.TypeSimpleString, model.TypeError:
		return parseSimpleStringOutput(result)
	case model.TypeBulkString:
		return parseBulkStringOutput(result)
	case model.TypeArray:
		return parseArrayOutput(result)
	}

	return "+OK\r\n"
}

func parseArrayOutput(result *model.RespOutput) string {

	data := result.Data.([]string)

	var resultBuilder strings.Builder
	length := len(data)

	if data[0] == "COMMAND" {
		return "+OK\r\n"
	}

	resultBuilder.WriteString(fmt.Sprintf("%s%d\r\n", string(model.TypeArray), length))

	for _, element := range data {
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

	if result == "-1" {
		return fmt.Sprintf("%s-1\r\n", string(data.RespType))
	}

	length := len(result)
	return fmt.Sprintf("%s%d\r\n%s\r\n", string(data.RespType), length, result)
}
