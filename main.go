package main

import (
	myCmd "github.com/ShadowFlade/observer/cmd"
	"github.com/ShadowFlade/observer/pkg/db"
	"github.com/ShadowFlade/observer/pkg/logic"
)


func main() {
	app := logic.App{DebugState: logic.DEBUG_DEBUG}
	db := db.Db{}
	if !db.IsDbPresent() {
		db.CreateSchema()
	}

	app.Main("shadowflade", 1, db)
	myCmd.Execute()
}
