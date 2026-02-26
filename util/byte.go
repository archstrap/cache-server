package util

import (
	"github.com/archstrap/cache-server/pkg/model"
	"github.com/archstrap/cache-server/pkg/parser"
)

func GetBytes(input *model.RespValue) int {
	return len(parser.ParseOutput(input.ToRespOutput()))
}
