package eventloop

import (
	"io"
	"log"
	"net"

	"github.com/archstrap/cache-server/internal/command"
	parserLib "github.com/archstrap/cache-server/pkg/parser"
)

type RedisTask struct {
	Connection net.Conn
}

func (conn *RedisTask) exec() {

	connection := conn.Connection

	for {
		data, err := parserLib.Parse(connection)

		if err != nil {

			if err == io.EOF {
				log.Printf("Client [ %s ] disconnected", connection.RemoteAddr())
				break
			}
			log.Fatalf("Error occurred. reason %v", err)
			break
		}

		factory := command.NewCommandHandlerFactory()
		output := factory.ProcessCommand(data)
		_, err = connection.Write([]byte(output))

		if err == nil {
			log.Println("Output flushed successfully")
		}

	}

}
