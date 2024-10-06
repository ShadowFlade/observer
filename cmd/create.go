package cmd

import (
	"github.com/ShadowFlade/observer/cmd/ui/textInput"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	logoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#01FAC6")).Bold(true)
	tipMsgStyle = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("190")).Italic(true)
)

func init() {
	rootCmd.AddCommand()
}

type listOptions struct {
	options []string
}

type Options struct {
	ProjectName textInput.Output
	ProjectType string
}

// var createCmd = &cobra.Command{
// 	Use:   "",
// 	Short: "short description",
// 	Long:  ".",
// 	Run: func(cmd *cobra.Command, args []string) {
//
// 	}
// }
