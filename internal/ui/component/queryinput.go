package component

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type QueryInput struct {
	Model textinput.Model
	Style lipgloss.Style
	// height    int
	// width     int
	// limit     int
}

func New() QueryInput {
	ti := textinput.New()
	ti.Placeholder = "Query"
	ti.Focus()
	ti.CharLimit = 256

	st := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	return QueryInput{
		Model: ti,
		Style: st,
	}
}

func (qi *QueryInput) Redraw(termHeight int, termWidth int) {
	// Magic numbers, should be remvoed when using lip gloss as container
	padding := 2
	overflowRight :=  1
	modelWidth := padding + len(qi.Model.Prompt) + overflowRight

	qi.Model.Width = termWidth - modelWidth
}
