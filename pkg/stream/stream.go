package stream

import (
	"context"
	"log/slog"
	"time"

	"github.com/augustomelo/stail/pkg/source"
)

type Stream struct {
	play  chan struct{}
	pause chan struct{}
	stop  chan struct{}
}

func (s *Stream) Play() {
	s.play <- struct{}{}
}

func (s *Stream) Pause() {
	s.pause <- struct{}{}
}

func (s *Stream) Stop() {
	s.stop <- struct{}{}
}

func (s *Stream) Start(ctx context.Context, src source.Source, dst chan source.Log) {
	go func(s *Stream) {
		shared := make(chan *[]byte)
		go src.Map(shared, dst)

	OutterLoop:
		for {
			select {
			case <-s.pause:
				for {
					slog.Debug("Paused producing logs")
					select {
					case <-s.play:
						slog.Debug("Resume producing logs")
					case <-s.stop:
						slog.Debug("Stop producing logs")
						break OutterLoop
					}
				}
			case <-s.stop:
				slog.Debug("Stop producing logs")
				break OutterLoop
			default:
				src.Produce(ctx, shared)
				time.Sleep(1 * time.Second)
			}
		}
		close(shared)
	}(s)
}
