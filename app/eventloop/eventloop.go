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

func (eventLoop *EventLoop) Start(shutDownSignal <-chan struct{}) {

	log.Println("EventLoop started...........")

loop:
	for {
		select {
		case redisTask := <-eventLoop.Tasks:
			go redisTask.execute()
		case <-shutDownSignal:
			break loop
		}
	}

	log.Println("EventLoop terminated")
}
