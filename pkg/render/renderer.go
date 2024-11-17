package render

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

var (
	logoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#01FAC6")).Bold(true)
	usersStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("190")).Italic(true).Width(8)
	memStyle   = lipgloss.NewStyle().PaddingLeft(1).Bold(true).Align(lipgloss.Right)
)

type Renderer struct {
}

func (r *Renderer) RenderUser(user string, totalMemUsage float32) {
	fmt.Printf(
		"%s %s\n",
		usersStyle.Render(string(user)),
		memStyle.Render(
			strconv.FormatFloat(float64(totalMemUsage), 'f', 2, 32)),
	)
}

