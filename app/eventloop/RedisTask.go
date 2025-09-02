package eventloop

import (
	"fmt"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

type RedisTask struct {
	Connection net.Conn
}

func (redisTask *RedisTask) execute() {
	connection := redisTask.Connection

	log.Println("Processing command from:", connection.RemoteAddr())

	for {

		buffer := make([]byte, 1024)
		n, err := connection.Read(buffer)

		if err != nil {
			log.Println("Error occurred. Details:", err)
			break
		}
		receivedMessage := string(buffer[:n])

		parser := parser.InputParser{InputMessage: receivedMessage}

		var outputMessage string
		result, err := parser.Parse()

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
