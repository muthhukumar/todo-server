package internal

import (
	"database/sql"
	"fmt"
	"time"
	"todo-server/models"

	"github.com/robfig/cron/v3"
)

func SetupCronJobs(db *sql.DB, emailAuth models.EmailAuth) {
	c := cron.New(cron.WithSeconds())

	c.AddFunc("*/60 * * * * *", func() {
		today := time.Now().Format("2006-01-02")

		query := "select name, due_date from tasks where due_date = $1 ORDER BY created_at DESC"

		rows, err := db.Query(query, today)

		if err != nil {
			fmt.Println("Failed to run the query", err.Error())
			return
		}

		var tasks []models.Task

		for rows.Next() {
			var task models.Task
			if err := rows.Scan(&task.Name, &task.DueDate); err != nil {
				fmt.Println("Failed to set task data")
				return
			}

			tasks = append(tasks, task)
		}
		defer rows.Close()

		var body = fmt.Sprintf("Tasks Due today: %v", today)

		body += "\n"
		for idx, task := range tasks {
			body += fmt.Sprintf("%d. %s", idx+1, task.Name)
			body += "\n"
		}

		template := models.EmailTemplate{
			To:      []string{emailAuth.ToEmail},
			Subject: "Today's Tasks",
			Body:    body,
		}

		email_sent := SendEmail(emailAuth, template)

		fmt.Println("Email send", email_sent, time.Now())
	})

	c.Start()

	fmt.Println("Cron jobs have been set up successfully.", time.Now())
}
