package command

import (
	"slices"
	"strconv"
	"strings"
	"time"

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

	startTime := time.Now()
	args := input.ArgsToStringSlice()
	n := len(args)

	streamIndex := slices.IndexFunc(args, func(s string) bool {
		return strings.ToUpper(s) == "STREAMS"
	})
	remaining := (n - 1) - streamIndex
	end := remaining / 2
	storage := store.StreamStoreInstance
	items := make([]*store.Pair[string, string], 0)
	dollarExists := slices.Contains(args, "$")

	for i := streamIndex + 1; i <= streamIndex+end; i++ {
		key := args[i]
		var id string

		if dollarExists {
			id = "$"
		} else {
			id = args[i+end]
		}

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
		result = storage.SearchExclusiveWithBlock(items, timeOut, dollarExists, startTime)
	}

	return model.NewRespOutput(model.TypeArray, result)
}
