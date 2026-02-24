package replication

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"

	"github.com/archstrap/cache-server/pkg/model"
	"github.com/archstrap/cache-server/pkg/parser"
)

type ReplicationDetails struct {
	conn net.Conn
	port string // used for logging monitoring purpose
}

func NewReplicationDetails(conn net.Conn, port string) *ReplicationDetails {
	return &ReplicationDetails{
		conn: conn,
		port: port,
	}
}

func (r *ReplicationDetails) GetTcpAddr() string {
	tcp := r.conn.RemoteAddr().(*net.TCPAddr)
	return fmt.Sprintf("%s:%d", tcp.IP, tcp.Port)
}

func (r *ReplicationDetails) Propagate(input *model.RespValue) {

	conn := r.conn
	command := model.NewRespOutput(input.DataType, input.Value)
	bytes := parser.ParseOutput(command)
	if _, err := conn.Write([]byte(bytes)); err != nil {
		if err == io.EOF {
			slog.Info("Connection Closed while sending the details to replica", slog.Any("Details", r.GetTcpAddr()))
			return
		}
		slog.Error("Unable to Send data to replica", slog.Any("input", input.String()))
		return
	}

	data, err := parser.Parse(conn)
	if err != nil {
		if err == io.EOF {
			slog.Info("Connection Closed while sending the details to replica", slog.Any("Details", r.GetTcpAddr()))
			return
		}
		slog.Error("Unable to Send data to replica", slog.Any("input", input.String()))
		return
	}

	slog.Info("", "Received from replica ", slog.Any("addr", r.GetTcpAddr()), slog.Any("out", data.String()))
}

type ReplicationStore struct {
	replications []*ReplicationDetails
	lock         sync.RWMutex
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

func (s *ReplicationStore) Add(conn net.Conn, port string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.replications = append(s.replications, NewReplicationDetails(conn, port))
	slog.Info("Added Replica Connection ", slog.Any("Details", port))
}

func (s *ReplicationStore) Propagate(input *model.RespValue) {
	if len(s.replications) == 0 {
		slog.Info("No Replications are there to send data")
		return
	}

	slog.Info("Propagation Initiated")

	for i := range s.replications {
		s.replications[i].Propagate(input)
	}
}

func (s *ReplicationStore) ActiveReplicationCount() int {
	return len(s.replications)
}

func (s *ReplicationStore) PrintState() {
	s.lock.RLock()
	defer s.lock.RUnlock()
	slog.Info("Active ReplicationStore", slog.Any("count", s.ActiveReplicationCount()))
	fmt.Println("Following Active Replications are listed below")
	for i := range s.replications {
		fmt.Printf("No-%d, Connection Details: %s, Port: %s\n", (i + 1), s.replications[i].GetTcpAddr(), s.replications[i].port)
	}
}
