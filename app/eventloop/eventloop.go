package eventloop

import "fmt"

type EventLoop struct {
	Tasks chan RedisTask
}

func (eventLoop *EventLoop) AddEvent(redisTask RedisTask) {
	eventLoop.Tasks <- redisTask
}

func (eventLoop *EventLoop) Start() {
	for {
		select {
		case redisTask := <-eventLoop.Tasks:
			fmt.Print("task is getting executed! ")
			go redisTask.execute()
		}
	}
}
