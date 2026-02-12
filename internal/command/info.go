package command

import (
	"fmt"
	"log/slog"
	"strings"

	// "strings"

	"github.com/archstrap/cache-server/internal/config"
	"github.com/archstrap/cache-server/pkg/model"
)

type ReplicaDetails map[string]string

var replication = make(ReplicaDetails)

func LoadServerWithMode() {

	if config.ReplicaOf == "" {
		LoadAsMaster()
		return
	}

	LoadAsReplica()

}

func LoadAsMaster() {

	replication["role"] = "master"
	replication["master_replid"] = getMasterReplicationId()
	replication["master_repl_offset"] = "0"
	slog.Info("server started as Master Node")
	slog.Info("", slog.Any("Details", getReplicaDetails()))
}

func LoadAsReplica() {

	// masterDetails := strings.Split(config.ReplicaOf, " ")
	replication["role"] = "slave"

	slog.Info("server started as Replica Node")
}

func getMasterReplicationId() string {
	return "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
}

func getReplicaDetails() string {

	result := make([]string, 0, len(replication))

	for k, v := range replication {
		result = append(result, fmt.Sprintf("%s:%s", k, v))
	}

	return fmt.Sprintf("%s\r\n", strings.Join(result, "\r\n"))

}

type InfoCommand struct {
}

var InfoCommandInstance = &InfoCommand{}

func (i *InfoCommand) Process(value *model.RespValue) *model.RespOutput {

	args := value.Value.([]string)

	if len(args) < 2 {
		return model.NewUnknownCommandOutput("ERR wrong number of arguments for 'info' command")
	}

	switch args[1] {
	case "replication":
		return model.NewRespOutput(model.TypeBulkString, getReplicaDetails())
	}

	return model.NewUnknownCommandOutput("ERR wrong number of arguments for 'info' command")
}

func (i *InfoCommand) Name() string {
	return "INFO"
}
