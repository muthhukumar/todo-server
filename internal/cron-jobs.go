package internal

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"text/tabwriter"
	"time"
	"todo-server/backup"
	data "todo-server/data/quotes"
	"todo-server/models"

	"github.com/robfig/cron/v3"
)

func SetupCronJobs(db *sql.DB, emailAuth models.EmailAuth) {
	c := cron.New(cron.WithSeconds())

	// Sunday evening 5 o clock
	c.AddFunc("0 0 17 * * 0", func() {
		backup.BackupTasks(db, emailAuth)
	})

	c.AddFunc("0 0 9 * * *", func() {

		quotes := data.GetRandomQuotes(data.GetQuotes(), 2)

		var body = fmt.Sprintf("Quotes of the day: %v", time.Now().Format("Monday, January 2 2006"))

		body += "\n"
		body += "\n"

		for idx, quote := range quotes {
			body += fmt.Sprintf("%d. %s", idx+1, quote)
			body += "\n"
			body += "\n"
		}

		template := models.EmailTemplate{
			To:      []string{emailAuth.ToEmail},
			Subject: "Quotes of the day",
			Body:    body,
		}

		email_sent := SendEmail(emailAuth, template)

		log.Println("Email send for quote of the day", email_sent, time.Now())
	})

	c.AddFunc("0 30 1 * * *", func() {
		today := time.Now().Format("2006-01-02")

		query := "select name, due_date from tasks where due_date = $1 and completed = false ORDER BY created_at DESC"

		rows, err := db.Query(query, today)

		if err != nil {
			log.Println("Failed to run the query", err.Error())
			return
		}

		var tasks []models.Task

		for rows.Next() {
			var task models.Task
			if err := rows.Scan(&task.Name, &task.DueDate); err != nil {
				log.Println("Failed to set task data")
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

		log.Println("Email send", email_sent, time.Now())
	})

	c.AddFunc("0 30 16 * * *", func() {
		// TODO: fix this completed on date issue. Probably have to default the value to empty string instead of null. Have to change the table schema
		query := "SELECT name FROM tasks WHERE completed = true AND DATE(NULLIF(completed_on, '')) = CURRENT_DATE;"

		rows, err := db.Query(query)

		if err != nil {
			log.Println("Failed to run the query", err.Error())
			return
		}

		var tasks []models.Task

		for rows.Next() {
			var task models.Task
			if err := rows.Scan(&task.Name); err != nil {
				log.Println("Failed to set task data")
				return
			}

			tasks = append(tasks, task)
		}
		defer rows.Close()

		totalTasksQuery := "select count(*) from tasks"

		var totalTasks int
		count_err := db.QueryRow(totalTasksQuery).Scan(&totalTasks)

		if count_err != nil {
			log.Println("Failed to run count query", count_err.Error())
		}

		totalCompletedTasksQuery := "select count(*) from tasks where completed = true;"

		var totalCompletedTasks int
		curr_err := db.QueryRow(totalCompletedTasksQuery).Scan(&totalCompletedTasks)

		if curr_err != nil {
			log.Println("Failed to run count query", curr_err.Error())
		}

		var body = fmt.Sprintf("Tasks completed Today: %v", time.Now().Format("Monday, January 2 2006"))

		body += "\n"
		body += "\n"
		for idx, task := range tasks {
			body += fmt.Sprintf("%d. %s", idx+1, task.Name)
			body += "\n"
		}

		body += "\n"
		body += getCompletedTasksTable(totalTasks, totalCompletedTasks)

		template := models.EmailTemplate{
			To:      []string{emailAuth.ToEmail},
			Subject: "Tasks completed Today",
			Body:    body,
		}

		email_sent := SendEmail(emailAuth, template)

		log.Println("Email send for completed tasks", email_sent, time.Now())
	})

	c.Start()

	log.Println("Cron jobs have been set up successfully.", time.Now())
}

func getCompletedTasksTable(totalTasks int, totalCompletedTasks int) string {
	var buf bytes.Buffer

	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', tabwriter.TabIndent)

	fmt.Fprintln(w, "ID\tTitle\tCount\t")

	// TODO: There are some warning here. Check that also.
	fmt.Fprintln(w, fmt.Sprintf("%d\t%s\t%d\t", 1, "Total Tasks", totalTasks))
	fmt.Fprintln(w, fmt.Sprintf("%d\t%s\t%d\t", 2, "Total Completed Tasks", totalCompletedTasks))

	w.Flush()

	result := buf.String()

	return result

}

// TODO: finish this. This is for the list of tasks for the day
// func getCompletedTasksTable(tasks []models.Task) string {
// 	var buf bytes.Buffer
//
// 	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', tabwriter.AlignRight)
//
// 	for idx, task := range tasks {
// 		fmt.Fprintln(w, "")
// 	}
//
// 	result := buf.String()
//
// 	return result
//
// }
