package util

import (
	"strings"

	"github.com/archstrap/cache-server/pkg/model"
)

func IsInputGetAck(input *model.RespValue) bool {

	if input.Command != "REPLCONF" {
		return false
	}

	args, ok := input.Value.([]string)
	if !ok {
		return false
	}

	if len(args) < 3 {
		return false
	}

	return strings.ToUpper(args[1]) == "GETACK"
}

func IsInputAck(input *model.RespValue) bool {

	if input.Command != "REPLCONF" {
		return false
	}

	args, ok := input.Value.([]string)
	if !ok {
		return false
	}

	if len(args) < 3 {
		return false
	}

	return strings.ToUpper(args[1]) == "ACK"
}
