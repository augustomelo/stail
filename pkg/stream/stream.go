package stream

import (
	"context"
	"log/slog"
	"time"

	"github.com/augustomelo/stail/pkg/source"
)

type Stream struct {
	pause  chan struct{}
	resume chan struct{}
	stop   chan struct{}
	src    source.Source
}

func (s *Stream) Pause() {
	slog.Debug("Pausing stream")
	s.pause <- struct{}{}
}

func (s *Stream) Resume() {
	slog.Debug("Resuming stream")
	s.resume <- struct{}{}
}

func (s *Stream) Stop() {
	slog.Debug("Stopping stream")
	s.stop <- struct{}{}
}

func (s *Stream) UpdateQuery(q string) {
	s.src.UpdateQuery(q)
}

func (s *Stream) Start(ctx context.Context, src source.Source, dst chan source.Log) {
	s.src = src
	s.resume = make(chan struct{})
	s.pause = make(chan struct{})
	s.stop = make(chan struct{})

	go func(s *Stream) {
		shared := make(chan *[]byte)
		go s.src.Map(shared, dst)

	MainLoop:
		for {
			select {
			case <-s.pause:
			PauseLoop:
				for {
					select {
					case <-s.pause:
					case <-s.resume:
						break PauseLoop
					case <-s.stop:
						break MainLoop
					}
				}

			case <-s.stop:
				break MainLoop

			case <-s.resume:
			default:
				src.Produce(ctx, shared)
				time.Sleep(1 * time.Second)
			}
		}

		close(shared)
	}(s)
}
