package eventloop

import (
	"log"
	"net"
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
		log.Println("Received Command:", receivedMessage)

		_, err = connection.Write([]byte("+OK\r\n"))

		if err != nil {
			log.Fatalln("Error occurred while trying to write commands. Error details:", err)
		}
	}
	
}
