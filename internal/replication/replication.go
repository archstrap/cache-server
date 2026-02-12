package replication

import (
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/archstrap/cache-server/internal/config"
	"github.com/archstrap/cache-server/pkg/model"
	"github.com/archstrap/cache-server/pkg/parser"
)

// State holds replication role and metadata (master_replid, master_repl_offset, etc.).
type State struct {
	details map[string]string
}

var state *State

// InitFromConfig sets replication mode from config (ReplicaOf) and initializes state.
// Call once at startup after config.ReadFlags().
func InitFromConfig() {
	if config.ReplicaOf == "" {
		initAsMaster()
		return
	}
	initAsReplica()
}

func initAsMaster() {
	state = &State{
		details: map[string]string{
			"role":               "master",
			"master_replid":      getMasterReplicationID(),
			"master_repl_offset": "0",
		},
	}
	slog.Info("server started as Master Node")
	slog.Info("replication details", slog.Any("details", FormatDetails()))
}

func initAsReplica() {
	state = &State{
		details: map[string]string{
			"role": "slave",
		},
	}
	slog.Info("server started as Replica Node")
	go connectToMaster()
}

func getMasterReplicationID() string {
	return "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
}

// FormatDetails returns replication info as a single string (Redis INFO format with \r\n).
// Used by INFO replication and logging.
func FormatDetails() string {
	if state == nil || len(state.details) == 0 {
		return ""
	}
	lines := make([]string, 0, len(state.details))
	for k, v := range state.details {
		lines = append(lines, fmt.Sprintf("%s:%s", k, v))
	}
	return strings.Join(lines, "\r\n") + "\r\n"
}

func connectToMaster() {

	masterNodeDetails := strings.ReplaceAll(config.Store["replicaof"], " ", ":")

	conn, err := net.Dial("tcp", masterNodeDetails)
	if err != nil {
		slog.Error("Unable to connect to master node.", slog.Any("address", masterNodeDetails))
		return
	}
	defer conn.Close()

	serializedPingCommand := parser.ParseOutput(model.NewRespOutput(model.TypeArray, []string{"PING"}))
	conn.Write([]byte(serializedPingCommand))

	response, err := parser.Parse(conn)
	if err != nil {
		slog.Error("Unable to get the response from replica")
		return
	}

	slog.Info("Received ", slog.Any("response", response.String()))

	slog.Info("Connected to Master Node. ", slog.Any("Address", masterNodeDetails))

}
