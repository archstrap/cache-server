package util

import (
	"strconv"

	"github.com/archstrap/cache-server/pkg/model"
)

func ParseFloat(input string) (float64, *model.RespOutput) {
	data, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return -1, model.NewRespOutput(model.TypeError, err.Error())
	}

	return data, nil
}

