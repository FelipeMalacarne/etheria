package engine

import (
	"context"
	"time"
)

type Loop struct {
	tickRate time.Duration
	onTick   func(tick int64, delta time.Duration)
}

func NewLoop(tickRate time.Duration, onTick func(tick int64, delta time.Duration)) *Loop {
	return &Loop{
		tickRate: tickRate,
		onTick:   onTick,
	}
}

func (l *Loop) Start(ctx context.Context) {
	if l.tickRate <= 0 {
		return
	}

	ticker := time.NewTicker(l.tickRate)
	defer ticker.Stop()

	var tick int64

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tick += 1
			if l.onTick != nil {
				l.onTick(tick, l.tickRate)
			}
		}
	}
}
