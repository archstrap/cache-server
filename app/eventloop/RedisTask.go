package eventloop

import (
	"fmt"
	"net"
)

type RedisTask struct {
	Connection net.Conn
}

func (redisTask *RedisTask) execute() {
	connection := redisTask.Connection

	fmt.Println("Processing command from:", connection.RemoteAddr())

	for {

		buffer := make([]byte, 1024)
		n, err := connection.Read(buffer)

		if err != nil {
			fmt.Errorf("Error occurred. Details:", err)
		}
		receivedMessage := string(buffer[:n])
		fmt.Println("Received Command: %s", receivedMessage)

		connection.Write([]byte("+OK\r\n"))
	}

}
