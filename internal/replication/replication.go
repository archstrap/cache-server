package replication

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/archstrap/cache-server/internal/config"
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
