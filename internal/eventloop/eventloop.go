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

	for task := range orDone(shutDownSignal, eventLoop.Tasks) {
		if redisTask, ok := task.(RedisTask); ok {
			go redisTask.exec()
		}
	}

	log.Println("EventLoop terminated")
}

func orDone(done <-chan struct{}, dataChannel <-chan RedisTask) chan interface{} {

	relayStreams := make(chan interface{})

	go func() {
		defer close(relayStreams)

		for {
			select {
			case <-done:
				return
			case data, ok := <-dataChannel:
				if !ok {
					return
				}

				select {
				case relayStreams <- data:
				case <-done:
					return
				}
			}
		}
	}()

	return relayStreams

}
