package eventloop

import (
	"context"
	"io"
	"log"
	"net"

	"github.com/archstrap/cache-server/internal/command"
	parserLib "github.com/archstrap/cache-server/pkg/parser"
)

type RedisTask struct {
	Connection net.Conn
	Context    context.Context
}

func (conn *RedisTask) exec() {

	connection := conn.Connection

	for {
		select {
		case <-conn.Context.Done():
			log.Println("Shutdown server")
			return

		default:
			data, err := parserLib.Parse(connection)
			if err != nil {
				if err == io.EOF {
					log.Printf("Client [ %s ] disconnected", connection.RemoteAddr())
					return
				}
				log.Fatalf("Error occurred. reason %v", err)
			}

			factory := command.GetCommandHandlerFactory()
			output := factory.ProcessCommand(data)
			_, err = connection.Write([]byte(output))

			if err == nil {
				log.Println("Output flushed successfully")
			}
		}

	}

}
