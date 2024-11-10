package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/augustomelo/stail/internal/ui/view"
	tea "github.com/charmbracelet/bubbletea"
)


func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	slog.Debug("Start")

	p := tea.NewProgram(
		view.InitialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		slog.Error("Error while instantiating view", "err", err)
		os.Exit(1)
	}

	slog.Debug("End")
}
