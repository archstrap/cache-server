package command

import (
	"log/slog"
	"slices"
	"strconv"
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
	storage := store.StreamStoreInstance
	items := make([]*store.Pair[string, string], 0)

	for i := streamIndex + 1; i <= streamIndex+end; i++ {
		key := args[i]
		id := args[i+end]

		var item *store.Pair[string, string] = store.NewPair(key, id)
		items = append(items, item)
	}

	var result []any

	blockIndex := slices.IndexFunc(args, func(s string) bool {
		return strings.ToUpper(s) == "BLOCK"
	})

	if blockIndex == -1 {
		result = storage.SearchExclusiveWithoutBlock(items)
	} else {
		timeOut, _ := strconv.Atoi(args[blockIndex+1])
		slog.Info("XREAD", slog.Any("timeOut", timeOut))
		result = storage.SearchExclusiveWithBlock(items, timeOut)
	}

	return model.NewRespOutput(model.TypeArray, result)
}
