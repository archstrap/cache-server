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
	case model.TypeInteger:
		return parseIntegerOutput(result)
	}

	return "+OK\r\n"
}

func parseIntegerOutput(result *model.RespOutput) string {
	data := result.Data.(int)
	return fmt.Sprintf("%s%d\r\n", string(result.RespType), data)
}

func parseStringArrayOutput(data []string) string {
	if len(data) == 0 {
		return "*0\r\n"
	}

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

func parseAnyArrayOutput(data []any) string {
	var result strings.Builder

	if data == nil {
		return "*-1\r\n"
	}

	result.WriteString(fmt.Sprintf("*%d\r\n", len(data)))

	for i := range data {
		switch innerData := data[i].(type) {
		case string:
			result.WriteString(parseBulkStringOutput(model.NewRespOutput(model.TypeBulkString, innerData)))
		case int:
			result.WriteString(parseIntegerOutput(model.NewRespOutput(model.TypeInteger, innerData)))
		case []string:
			result.WriteString(parseStringArrayOutput(innerData))
		case []any:
			result.WriteString(parseAnyArrayOutput(innerData)) // recursively
		}
	}

	return result.String()
}

func parseArrayOutput(result *model.RespOutput) string {

	switch data := result.Data.(type) {
	case []string:
		return parseStringArrayOutput(data)
	case []any:
		return parseAnyArrayOutput(data)
	}

	return "*0\r\n"
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
