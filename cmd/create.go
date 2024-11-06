package cmd

import (
	"github.com/ShadowFlade/observer/cmd/ui/textInput"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	tipMsgStyle = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("190")).Italic(true)
)

func init() {
	rootCmd.AddCommand(createCmd)
}

type listOptions struct {
	options []string
}

type Options struct {
	ProjectName textInput.Output
	ProjectType string
}

var createCmd = &cobra.Command{
	Use:   "user",
	Short: "Get info about specific user",
	Long:  ".",
	Run: func(cmd *cobra.Command, args []string) {

	}}
