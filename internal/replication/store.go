package replication

import (
	"io"
	"log/slog"
	"net"
	"sync"

	"github.com/archstrap/cache-server/pkg/model"
	"github.com/archstrap/cache-server/pkg/parser"
)

type ReplicationDetails struct {
	addr string
}

func NewReplicationDetails(addr string) *ReplicationDetails {
	return &ReplicationDetails{
		addr: addr,
	}
}

func (r *ReplicationDetails) Propagate(input *model.RespValue) {
	conn, err := net.Dial("tcp", r.addr)
	if err != nil {
		slog.Error("Unable to Send Values", slog.Any("Details", err))
		return
	}

	defer conn.Close()
	command := model.NewRespOutput(input.DataType, input.Value)
	bytes := parser.ParseOutput(command)
	if _, err := conn.Write([]byte(bytes)); err != nil {
		if err == io.EOF {
			slog.Info("Connection Closed while sending the details to replica", slog.Any("Details", r.addr))
			return
		}
		slog.Error("Unable to Send data to replica", slog.Any("input", input.String()))
		return
	}

	data, err := parser.Parse(conn)
	if err != nil {
		if err == io.EOF {
			slog.Info("Connection Closed while sending the details to replica", slog.Any("Details", r.addr))
			return
		}
		slog.Error("Unable to Send data to replica", slog.Any("input", input.String()))
		return
	}

	slog.Info("", "Received from replica ", slog.Any("addr", r.addr), slog.Any("out", data.String()))
}

type ReplicationStore struct {
	replications []*ReplicationDetails
}

var (
	lock              sync.Mutex
	replicationStores *ReplicationStore
)

func InitReplicationStore() {
	if replicationStores == nil {

		if replicationStores == nil {
			lock.Lock()
			defer lock.Unlock()

			replicationStores = &ReplicationStore{
				replications: make([]*ReplicationDetails, 0),
			}

		}
	}

}

func GetReplicationStore() *ReplicationStore {
	InitReplicationStore()
	return replicationStores
}

func (s *ReplicationStore) AddReplication(addr string) {
	lock.Lock()
	defer lock.Unlock()
	s.replications = append(s.replications, NewReplicationDetails(addr))
}

func (s *ReplicationStore) Propagate(input *model.RespValue) {
	if len(s.replications) == 0 {
		slog.Info("No Replications are there to send data")
		return
	}
	for i := range s.replications {
		s.replications[i].Propagate(input)
	}
}
