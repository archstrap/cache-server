package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/archstrap/cache-server/internal/config"
	"github.com/archstrap/cache-server/internal/replication"
)

func MonitorReplicaConnections(ctx context.Context, duration time.Duration) {

	if config.Store["replicaof"] != "" {
		slog.Info("In replica mode. No need to monitor")
		return
	}

	slog.Info("In master mode. start monitoring")

	timer := time.NewTicker(duration)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Monitoring of Replica Connection is Stopped")
			return
		case <-timer.C:
			store := replication.GetReplicationStore()
			store.PrintState()
		}
	}
}
