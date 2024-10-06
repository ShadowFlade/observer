package main

import (
	myCmd "github.com/ShadowFlade/observer/cmd"
	"github.com/ShadowFlade/observer/pkg/logic"
)

func main() {
	app := logic.App{DebugState: logic.DEBUG_DEBUG}
	app.Main("shadowflade")
	myCmd.Execute()
}
