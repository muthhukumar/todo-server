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

	c.AddFunc("0 30 1 * * *", func() {
		today := time.Now().Format("2006-01-02")

		query := "select name, due_date from tasks where due_date = $1 and completed = false ORDER BY created_at DESC"

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

		var body = fmt.Sprintf("Today Task's: %v", time.Now().Format("Monday, January 2 2006"))

		body += "\n"
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

	c.AddFunc("0 30 16 * * *", func() {
		query := "SELECT name FROM tasks WHERE completed = true AND DATE(completed_on) = CURRENT_DATE;"

		rows, err := db.Query(query)

		if err != nil {
			fmt.Println("Failed to run the query", err.Error())
			return
		}

		var tasks []models.Task

		for rows.Next() {
			var task models.Task
			if err := rows.Scan(&task.Name); err != nil {
				fmt.Println("Failed to set task data")
				return
			}

			tasks = append(tasks, task)
		}
		defer rows.Close()

		totalTasksQuery := "select count(*) from tasks"

		var totalTasks int
		count_err := db.QueryRow(totalTasksQuery).Scan(&totalTasks)

		if count_err != nil {
			fmt.Println("Failed to run count query", count_err.Error())
		}

		totalCompletedTasksQuery := "select count(*) from tasks where completed = true;"

		var totalCompletedTasks int
		curr_err := db.QueryRow(totalCompletedTasksQuery).Scan(&totalCompletedTasks)

		if curr_err != nil {
			fmt.Println("Failed to run count query", curr_err.Error())
		}

		var body = fmt.Sprintf("Tasks completed Today: %v", time.Now().Format("Monday, January 2 2006"))

		body += "\n"
		body += "\n"
		for idx, task := range tasks {
			body += fmt.Sprintf("%d. %s", idx+1, task.Name)
			body += "\n"
		}

		body += "\n"
		body += fmt.Sprintf("Total Tasks          : %v\n", totalTasks)
		body += fmt.Sprintf("Total Completed Tasks: %v", totalCompletedTasks)

		template := models.EmailTemplate{
			To:      []string{emailAuth.ToEmail},
			Subject: "Tasks completed Today",
			Body:    body,
		}

		email_sent := SendEmail(emailAuth, template)

		fmt.Println("Email send for completed tasks", email_sent, time.Now())
	})

	c.Start()

	fmt.Println("Cron jobs have been set up successfully.", time.Now())
}
