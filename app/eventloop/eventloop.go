package eventloop

import (
	"log"
)

type EventLoop struct {
	Tasks chan RedisTask
}

func (eventLoop *EventLoop) AddEvent(redisTask RedisTask) {
	eventLoop.Tasks <- redisTask
}

func (eventLoop *EventLoop) Start() {

	log.Println("EventLoop started...........")

	for {
		select {
		case redisTask := <-eventLoop.Tasks:
			go redisTask.execute()
		}
	}
}
