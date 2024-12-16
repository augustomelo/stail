package view

import (
	"context"
	"fmt"

	"github.com/augustomelo/stail/internal/ui/component"
	"github.com/augustomelo/stail/pkg/source"
	"github.com/augustomelo/stail/pkg/stream"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	qi component.QueryInput
	tl component.TableLog
	s  *stream.Stream
	l  chan source.Log
}

func InitialModel() model {
	return model{
		qi: component.NewQueryInput(),
		tl: component.NewTableLog(),
		s:  &stream.Stream{},
		l:  make(chan source.Log),
	}
}

func startStream(m model) tea.Cmd {
	m.s.Start(context.Background(), source.BuildDataDogSource(), m.l)

	return nil
}

func waitForLogs(m model) tea.Cmd {
	return func() tea.Msg {
		return <-m.l
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		startStream(m),
		waitForLogs(m),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case source.Log:
		m.tl.UpdateRowLog(msg)
		return m, waitForLogs(m)

	case tea.KeyMsg:

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "i":
			if m.qi.Model.Focused() {
				break
			}

			m.tl.Model.Blur()
			m.qi.Model.Focus()
			m.s.Pause()
		case "esc", "enter":
			m.tl.Model.Focus()
			m.qi.Model.Blur()
			m.s.UpdateQuery(m.qi.Model.Value())
			m.s.Resume()
		}

	case tea.WindowSizeMsg:
		m.qi.RedrawQueryInput(msg.Height, msg.Width)
		m.tl.RedrawTableLog(msg.Height, msg.Width)
	}

	m.qi.Model, cmd = m.qi.Model.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.qi.Style.Render(m.qi.Model.View()),
		m.tl.Style.Render(m.tl.Model.View()),
	) + "\n\n"
}
