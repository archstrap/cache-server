package eventloop

import (
	"context"
	"log"
)

type EventLoop struct {
	Tasks chan RedisTask
}

func (eventLoop *EventLoop) AddEvent(redisTask RedisTask) {
	eventLoop.Tasks <- redisTask
}

func (eventLoop *EventLoop) Start(ctx context.Context) {

	log.Println("EventLoop started")

	for task := range orDone(ctx, eventLoop.Tasks) {
		if redisTask, ok := task.(RedisTask); ok {
			go redisTask.exec()
		}
	}

	log.Println("EventLoop terminated")
}

func orDone(ctx context.Context, dataChannel <-chan RedisTask) chan interface{} {

	relayStreams := make(chan any)

	go func() {
		defer close(relayStreams)

		for {
			select {
			case <-ctx.Done():
				return
			case data, ok := <-dataChannel:
				if !ok {
					return
				}

				select {
				case relayStreams <- data:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return relayStreams

}
