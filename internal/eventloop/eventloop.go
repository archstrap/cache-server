package eventloop

import (
	"context"
	"log/slog"

	"github.com/archstrap/cache-server/util"
)

type EventLoop struct {
	Tasks chan RedisTask
}

func (eventLoop *EventLoop) AddEvent(redisTask RedisTask) {
	eventLoop.Tasks <- redisTask
}

func (eventLoop *EventLoop) Start(ctx context.Context) {

	slog.Info("EventLoop started")

	for task := range util.OrDone(ctx, eventLoop.Tasks) {
		if redisTask, ok := task.(RedisTask); ok {
			go redisTask.exec()
		}
	}

	slog.Info("EventLoop terminated")
}
