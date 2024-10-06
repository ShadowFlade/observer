/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ShadowFlade/observer/pkg/logic"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

const logo = `
$$$$$$\  $$$$$$$\   $$$$$$\  $$$$$$$$\ $$$$$$$\  $$\    $$\ $$$$$$$$\ $$$$$$$\
$$  __$$\ $$  __$$\ $$  __$$\ $$  _____|$$  __$$\ $$ |   $$ |$$  _____|$$  __$$\
$$ /  $$ |$$ |  $$ |$$ /  \__|$$ |      $$ |  $$ |$$ |   $$ |$$ |      $$ |  $$ |
$$ |  $$ |$$$$$$$\ |\$$$$$$\  $$$$$\    $$$$$$$  |\$$\  $$  |$$$$$\    $$$$$$$  |
$$ |  $$ |$$  __$$\  \____$$\ $$  __|   $$  __$$<  \$$\$$  / $$  __|   $$  __$$<
$$ |  $$ |$$ |  $$ |$$\   $$ |$$ |      $$ |  $$ |  \$$$  /  $$ |      $$ |  $$ |
 $$$$$$  |$$$$$$$  |\$$$$$$  |$$$$$$$$\ $$ |  $$ |   \$  /   $$$$$$$$\ $$ |  $$ |
 \______/ \_______/  \______/ \________|\__|  \__|    \_/    \________|\__|  \__|
`

var (
	logoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#01FAC6")).Bold(true)
	usersStyle = lipgloss.NewStyle().PaddingLeft(10).Foreground(lipgloss.Color("190")).Italic(true)
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "observer",
	Short: "controll your users and observer their RAM usage",
	Run: func(cmd *cobra.Command, args []string) {
		app := logic.App{DebugState: logic.DEBUG_DEBUG}
		users, _ := app.Main("")
		fmt.Printf("%s\n", logoStyle.Render(logo))
		fmt.Printf("%s\n", usersStyle.Render(strings.Join(users, "\n")))
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.observer.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
