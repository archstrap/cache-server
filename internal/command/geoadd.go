package command

import (
	"fmt"

	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
	"github.com/archstrap/cache-server/util"
)

type GEOADD struct{}

func (g *GEOADD) Name() string { return "GEOADD" }

func (g *GEOADD) Process(value *model.RespValue) *model.RespOutput {
	args := value.ArgsToStringSlice()
	if len(args) < 5 {
		return model.NewWrongNumberOfOutput(g.Name())
	}

	longitude, err := util.ParseFloat(args[2])
	if err != nil {
		return err
	}

	latitude, err := util.ParseFloat(args[3])
	if err != nil {
		return err
	}

	if longitude < -180 || longitude > 180 {
		return model.NewRespOutput(model.TypeError, fmt.Sprintf("ERR invalid longitude,latitude pair %g,%g", longitude, latitude))
	}

	if latitude < -85.05112878 || latitude > 85.05112878 {
		return model.NewRespOutput(model.TypeError, fmt.Sprintf("ERR invalid longitude,latitude pair %g,%g", longitude, latitude))
	}

	key := args[1]
	score := fmt.Sprintf("%d", util.Encode(latitude, longitude))
	member := args[4]

	item := store.NewSetItem(key, score, member)
	bucket := store.GetSkipListBucket()

	return model.NewRespOutput(model.TypeInteger, bucket.Insert(item))
}

func init() {
	Registry.RegisterCommand(&GEOADD{})
}
