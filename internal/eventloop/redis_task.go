package eventloop

import (
	"fmt"
	"log"
	"net"

	"github.com/archstrap/cache-server/internal/parser"
	parserLib "github.com/archstrap/cache-server/pkg/parser"
)

type RedisTask struct {
	Connection net.Conn
}

func (conn *RedisTask) execute() {
	connection := conn.Connection

	log.Println("Processing command from:", connection.RemoteAddr())

	for {

		buffer := make([]byte, 1024)
		n, err := connection.Read(buffer)

		if err != nil {
			log.Println("Error occurred. Details:", err)
			break
		}
		receivedMessage := string(buffer[:n])

		inputParser := parser.InputParser{InputMessage: receivedMessage}

		var outputMessage string
		result, err := inputParser.Parse()

		if err != nil {
			outputMessage = fmt.Sprintf("-%s\r\n", err.Error())
		} else {
			outputMessage = fmt.Sprintf("+%s\r\n", result)
		}
		_, err = connection.Write([]byte(outputMessage))

		if err != nil {
			log.Fatalln("Error occurred while trying to write commands. Error details:", err)
		}
	}

}

func (conn *RedisTask) exec() {

	connection := conn.Connection

	for {
		data, err := parserLib.Parse(connection)

		if err != nil {
			log.Fatalf("Error occurred. reason %v", err)
		}

		output := parserLib.ParseOutput(data)
		_, err = connection.Write([]byte(output))

		if err == nil {
			log.Println("Output flushed successfully")
		}

	}

}
