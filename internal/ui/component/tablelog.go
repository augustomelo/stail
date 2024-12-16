package component

import (
	"log/slog"
	"time"

	"github.com/augustomelo/stail/pkg/source"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type TableLog struct {
	Model table.Model
	Style lipgloss.Style
	// Colums []table.Column
	// Rows   []table.Row
}

func NewTableLog() TableLog {
	columns := []table.Column{
		{Title: "Date", Width: 20},
		{Title: "Level", Width: 5},
		{Title: "Message", Width: 70},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	st := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	return TableLog{
		Model: t,
		Style: st,
	}
}

func (tl *TableLog) RedrawTableLog(termHeight int, termWidth int) {
	// Magic numbers, should be remvoed when using lip gloss as container
	paddingTable := 2
	heightTextInput := 7
	usableWidithTable := termWidth - paddingTable
	usableHeightTable := termHeight - paddingTable - heightTextInput

	tl.Model.SetWidth(usableWidithTable)
	tl.Model.SetHeight(usableHeightTable)
}

func (tl *TableLog) UpdateRowLog(log source.Log) {
	rows := append([]table.Row{}, table.Row{
		log.Timestamp.Format(time.DateTime),
		log.Level,
		log.Message,
	})

	// I think that I still need to deal with overflow of elements, so I don't
	// have a huge number of things not being displayed.
	slog.Debug("Updating table log", "row", rows)
	tl.Model.SetRows(append(rows, tl.Model.Rows()...))
}
