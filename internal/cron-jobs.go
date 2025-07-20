package internal

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"
	"todo-server/backup"
	data "todo-server/data/quotes"
	"todo-server/db"
	templates "todo-server/internal/templates/today-tasks"
	"todo-server/models"
	"todo-server/utils"

	"github.com/chromedp/chromedp"
	"github.com/robfig/cron/v3"
)

func sendQuotes(emailAuth models.EmailAuth) {
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
}

// ANSI color codes
const (
	reset   = "\033[0m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
)

func SyncURLTitle(dc *sql.DB) {
	urlTitles, err := db.GetAllURLTitles(dc)

	log.Println("Syncing URL Titles...")

	if urlTitles == nil {
		log.Printf("No urls found")

		return
	}

	if err != nil {
		log.Println("Failed to sync url titles because", err)
		return
	}

	chromePath := os.Getenv("CHROME_PATH")

	utils.Assert(chromePath != "", "Chrome Path ENV value is not set")

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(chromePath),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	for _, urlTitle := range urlTitles {
		if urlTitle.IsValid && urlTitle.Title != "" {
			continue
		}

		var pageTitle string

		log.Printf("Syncing %s", urlTitle.URL)

		err := chromedp.Run(ctx,
			chromedp.Navigate(urlTitle.URL),
			chromedp.WaitReady("body"),
			chromedp.Evaluate(`document.title`, &pageTitle),
		)

		if err != nil && pageTitle == "" {
			log.Printf("Failed to fetch title. Got %s%s%s\n", reset, err.Error(), reset)
			continue
		}

		if urlTitle.Title == pageTitle {
			log.Printf("`%s` and `%s` are same\n", urlTitle.Title, pageTitle)
			continue
		}

		if pageTitle == "" {
			log.Printf("%s%s%s URL title not available. Got empty", red, urlTitle.URL, reset)
			continue
		}

		log.Printf("%sSaving new title%s: %s`%s`%s for URL: %s%s%s\n", magenta, reset, green, pageTitle, reset, blue, urlTitle.URL, reset)

		_ = db.SaveOrUpdateURLTitle(dc, pageTitle, urlTitle.URL, true)

	}

	// TODO: log time it took to complete this action.
	log.Println("Syncing completed.")

}

func SetupCronJobs(db *sql.DB, emailAuth models.EmailAuth) {
	istLocation, _ := time.LoadLocation("Asia/Kolkata")

	c := cron.New(cron.WithSeconds(), cron.WithLocation(istLocation))

	// Every day morning 2:00
	c.AddFunc("0 0 2 * * *", func() {
		_, err := db.Query("TRUNCATE TABLE log")

		if err != nil {
			log.Println("Failed to delete logs from db")

			return
		}

		log.Println("Deleted logs successfully")
	})

	// c.AddFunc("0 0 0 * * *", func() {
	// 	SyncURLTitle(db)
	// })

	// Every day morning 7:00 AM
	c.AddFunc("0 0 7 * * *", func() {
		sendQuotes(emailAuth)
	})

	// Every day morning 3:00 AM
	c.AddFunc("0 0 3 * * *", func() {
		backup.BackupTasks(db, emailAuth)
	})

	c.AddFunc("0 0 9 * * *", func() {
		sendQuotes(emailAuth)
	})

	// Today's Tasks. Every morning 7:00 AM
	c.AddFunc("0 0 7 * * *", func() {
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

		var bodyBuff bytes.Buffer

		tl, err := templates.TodayTasksEmailTemplate()

		if err != nil {
			log.Println("Failed to generate the template for email", err.Error())
			return
		}

		tl.Execute(&bodyBuff, tasks)

		msg := "From: " + emailAuth.FromEmail + "\n" +
			"To: " + emailAuth.ToEmail + "\n" +
			"Subject: Today's Task List\n" +
			"MIME-version: 1.0;\n" +
			"Content-Type: text/html; charset=\"UTF-8\";\n\n" +
			bodyBuff.String()

		email_sent := SendHtmlEmail(emailAuth, []string{emailAuth.ToEmail}, []byte(msg))

		log.Println("Email send", email_sent, time.Now())
	})

	// Everyday night 10:00 PM
	c.AddFunc("0 0 22 * * *", func() {
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

	// for _, entry := range c.Entries() {
	// 	c.Remove(entry.ID)
	// }

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
