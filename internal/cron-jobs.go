package internal

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

func SetupCronJobs(db *sql.DB) {
	c := cron.New()

	c.AddFunc("* * * * *", func() {
		fmt.Println("Running cron job: ", time.Now())
	})

	c.Start()

	fmt.Println("Cron jobs have been set up successfully.")
}
