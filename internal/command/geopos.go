package command

import (
	"fmt"
	"strconv"

	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
	"github.com/archstrap/cache-server/util"
)

type GEOPOS struct{}

// Name implements ICommand.
func (g *GEOPOS) Name() string {
	return "GEOPOS"
}

// Process implements ICommand.
func (g *GEOPOS) Process(value *model.RespValue) *model.RespOutput {
	args := value.ArgsToStringSlice()
	if len(args) < 3 {
		return model.NewWrongNumberOfOutput(g.Name())
	}

	key := args[1]

	bucket := store.GetSkipListBucket()
	// if !bucket.Exists(key) {
	// 	return model.NewRespOutput(model.TypeArray, nil)
	// }
	//
	var result []any

	for i := 2; i < len(args); i++ {
		member := args[i]
		score := bucket.Score(key, member)

		if score == "-1" {
			result = append(result, nil)
			continue
		}

		scores, err := strconv.ParseFloat(score, 64)
		if err != nil {
			return model.NewRespOutput(model.TypeError, err.Error())
		}

		coordinates := util.Decode(uint64(scores))

		result = append(result, []any{fmt.Sprintf("%g", coordinates.Longitude), fmt.Sprintf("%g", coordinates.Latitude)})
	}
	return model.NewRespOutput(model.TypeArray, result)

}

func init() {
	Registry.RegisterCommand(&GEOPOS{})
}
