package util

import (
	"context"
)

func OrDone[K any](ctx context.Context, dataChannel <-chan K) chan any {

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
