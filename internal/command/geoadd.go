package command

import (
	"fmt"

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

	lattitude, err := util.ParseFloat(args[3])
	if err != nil {
		return err
	}

	if longitude < -180 || longitude > 180 {
		return model.NewRespOutput(model.TypeError, fmt.Sprintf("ERR invalid longitude,latitude pair %g,%g", longitude, lattitude))
	}

	if lattitude < -85.05112878 || lattitude > 85.05112878 {
		return model.NewRespOutput(model.TypeError, fmt.Sprintf("ERR invalid longitude,latitude pair %g,%g", longitude, lattitude))
	}

	return model.NewRespOutput(model.TypeInteger, 1)
}

func init() {
	Registry.RegisterCommand(&GEOADD{})
}
