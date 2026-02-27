package command

import (
	"slices"
	"strings"

	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
)

type XREAD struct{}

var cXREAD = &XREAD{}

func init() {
	Registry.RegisterCommand(cXREAD)
}

func (x *XREAD) Name() string {
	return "XREAD"
}

func (x *XREAD) Process(input *model.RespValue) *model.RespOutput {

	args := input.ArgsToStringSlice()
	n := len(args)
	streamIndex := slices.IndexFunc(args, func(s string) bool {
		return strings.ToUpper(s) == "STREAMS"
	})
	remaining := (n - 1) - streamIndex
	end := remaining / 2

	result := make([]any, 0)

	storage := store.StreamStoreInstance

	for i := streamIndex + 1; i <= streamIndex+end; i++ {
		key := args[i]
		id := args[i+end]

		nested := storage.SearchExclusive(key, id)

		data := []any{
			key,
			nested,
		}

		result = append(result, data)

	}

	return model.NewRespOutput(model.TypeArray, result)
}
