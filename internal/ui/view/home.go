package view

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/augustomelo/stail/internal/ui/component"
	"github.com/augustomelo/stail/pkg/source"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	qi component.QueryInput
	tl component.TableLog
}

func InitialModel() model {
	return model{
		qi: component.NewQueryInput(),
		tl: component.NewTableLog(),
	}
}

func fetchLogs() tea.Msg {
	source := source.BuildDataDogSource()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	body, err := source.Produce(ctx)
	if err != nil {
		slog.Error("Error while producing logs", "err", err)
	}

	return source.Map(body)
}

func (m model) Init() tea.Cmd {
	return fetchLogs
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case []source.Log:
		m.tl.UpdateRowLog(msg)

	case tea.KeyMsg:

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
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
