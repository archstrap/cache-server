package command

import (
	"fmt"
	"strconv"

	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
	"github.com/archstrap/cache-server/util"
)

type GEODIST struct{}

// Name implements ICommand.
func (g *GEODIST) Name() string {
	return "GEODIST"
}

// Process implements ICommand.
func (g *GEODIST) Process(value *model.RespValue) *model.RespOutput {

	args := value.ArgsToStringSlice()
	if len(args) < 4 {
		return model.NewWrongNumberOfOutput(g.Name())
	}

	key := args[1]
	member1 := args[2]
	member2 := args[3]
	bucket := store.GetSkipListBucket()

	score1 := bucket.Score(key, member1)
	score2 := bucket.Score(key, member2)

	scores1, err := strconv.ParseFloat(score1, 64)
	if err != nil {
		return model.NewRespOutput(model.TypeError, err.Error())
	}
	scores2, err := strconv.ParseFloat(score2, 64)
	if err != nil {
		return model.NewRespOutput(model.TypeError, err.Error())
	}

	coOrdinates1 := util.Decode(uint64(scores1))
	coOrdinates2 := util.Decode(uint64(scores2))

	return model.NewRespOutput(model.TypeBulkString, fmt.Sprintf("%.4f", coOrdinates1.GetDistance(&coOrdinates2)))
}

func init() {
	Registry.RegisterCommand(&GEODIST{})
}
