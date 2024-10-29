package component

import (
	"time"

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
		{Title: "Date", Width: 4},
		{Title: "Level", Width: 10},
		{Title: "Message", Width: 10},
	}

	rows := []table.Row{
		{time.Now().String(), "INFO", "Something happened"},
		{time.Now().Add(-time.Duration(40) * time.Minute).String(), "ERROR", "A wild error appear"},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
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

func (tl *TableLog) RedrawTableLog(termHeight int, termWidth int)  {
	// Magic numbers, should be remvoed when using lip gloss as container
	paddingTable := 2
	heightTextInput := 7
	usableWidithTable := termWidth - paddingTable
	usableHeightTable := termHeight - paddingTable - heightTextInput

	tl.Model.SetWidth(usableWidithTable)
	tl.Model.SetHeight(usableHeightTable)
}
