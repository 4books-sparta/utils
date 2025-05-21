package observer

import (
	"context"
	"time"
)

type command int

const (
	commandFlush command = iota
	commandFlushAndWait
	commandFlushDone
)

const (
	defaultTickerPeriod = 1 * time.Second
)

type handler[T any] struct {
	queue        *queue[T]
	fn           EventHandler[T]
	commandCh    chan command
	tickerPeriod time.Duration
}

func newHandler[T any](queue *queue[T], fn EventHandler[T]) *handler[T] {
	return &handler[T]{
		queue:        queue,
		fn:           fn,
		commandCh:    make(chan command),
		tickerPeriod: defaultTickerPeriod,
	}
}

func (h *handler[T]) withTick(period time.Duration) *handler[T] {
	h.tickerPeriod = period
	return h
}

func (h *handler[T]) listen(ctx context.Context) {
	ticker := time.NewTicker(h.tickerPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			close(h.commandCh)
			return
		case <-ticker.C:
			handleCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			h.handle(handleCtx)
			cancel()
		case cmd, ok := <-h.commandCh:
			if !ok {
				return
			}

			handleCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			h.handle(handleCtx)
			cancel()

			if cmd == commandFlushAndWait {
				ticker.Stop()
				close(h.commandCh)
				return
			}
		}
	}
}

func (h *handler[T]) handle(ctx context.Context) {
	events := h.queue.All()
	if len(events) > 0 {
		h.fn(ctx, events)
	}
}

func (h *handler[T]) flush() {
	select {
	case h.commandCh <- commandFlush:
	default:
		// se il channel è bloccato, skip
	}
}

func (h *handler[T]) flushAndWait() {
	select {
	case h.commandCh <- commandFlushAndWait:
		// se il channel non è chiuso, aspetta
		_, ok := <-h.commandCh
		if !ok {
			return
		}
	default:
		// Se il channel è bloccato, è gia terminato
	}
}
