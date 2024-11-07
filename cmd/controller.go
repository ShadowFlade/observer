package cmd

import (
	"fmt"
	"time"

	"github.com/ShadowFlade/observer/pkg/db"
	"github.com/ShadowFlade/observer/pkg/logic"
)

type Controller struct {
}

type UserName interface {
	String() string
}

func (c *Controller) Start(user string, intervalSeconds int) {
	app := logic.App{DebugState: logic.DEBUG_DEBUG}
	db := db.Db{}
	regularUsers, ids := db.GetRegularUsers()
	onlyUser := app.FormatUsernameTop(user)
	interval := intervalSeconds * int(time.Second)
	ticker := time.NewTicker(time.Duration(interval))
	done := make(chan bool)
	fmt.Println("started ticker")
	defer ticker.Stop()

	go app.Main(onlyUser, intervalSeconds, db, regularUsers, ids)

	done <- true
	fmt.Println("Ticker done")
}
