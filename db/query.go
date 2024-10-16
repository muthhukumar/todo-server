package db

import (
	"database/sql"
	"fmt"
	"log"
	"todo-server/models"
)

func SaveOrUpdateURLTitle(db *sql.DB, title, url string, isValid bool) error {
	query := `
    INSERT INTO url_titles (title, url, is_valid)
    VALUES ($1, $2, $3)
    ON CONFLICT (url)
    DO UPDATE SET 
        title = EXCLUDED.title,
        is_valid = EXCLUDED.is_valid;
    `

	err := db.QueryRow(query, title, url, isValid).Err()
	if err != nil {
		log.Printf("Error saving or updating URL title: %v", err)
		return err
	}
	return nil
}

func GetURLTitle(db *sql.DB, url string) (*models.URLTitle, error) {
	query := "SELECT title, is_valid, url FROM url_titles WHERE url = $1"

	var urlTitle models.URLTitle
	row := db.QueryRow(query, url)

	err := row.Scan(&urlTitle.Title, &urlTitle.IsValid, &urlTitle.URL)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No entry found for the provided URL")
			return nil, nil // No result, but not an error
		}
		log.Printf("Error retrieving URL title: %v", err)
		return nil, err // Return the error if something went wrong
	}

	return &urlTitle, nil
}

func GetAllURLTitles(db *sql.DB) ([]models.URLTitle, error) {
	query := "SELECT title, is_valid, url FROM url_titles"

	rows, err := db.Query(query)

	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var urlTitles []models.URLTitle

	for rows.Next() {
		var urlTitle models.URLTitle

		err := rows.Scan(&urlTitle.Title, &urlTitle.IsValid, &urlTitle.URL)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		urlTitles = append(urlTitles, urlTitle)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		return nil, err
	}

	return urlTitles, nil
}

func ToggleTaskAndHandleRecurrence(db *sql.DB, taskID int) error {
	query := `
	WITH updated_task AS (
    UPDATE tasks 
    SET completed = NOT completed, 
        completed_on = CASE 
            WHEN completed = FALSE THEN TO_CHAR(CURRENT_TIMESTAMP, 'YYYY-MM-DD HH24:MI:SS')
            ELSE '' 
        END 
    WHERE id = $1
    RETURNING id, name, completed, recurrence_pattern, recurrence_interval, start_date, due_date
)

INSERT INTO tasks (name, start_date, due_date, recurrence_pattern, recurrence_interval)
SELECT 
    ut.name, 
    CASE
        WHEN ut.recurrence_pattern = 'daily' THEN TO_CHAR(CURRENT_TIMESTAMP + ut.recurrence_interval * INTERVAL '1 day', 'YYYY-MM-DD')
        WHEN ut.recurrence_pattern = 'weekly' THEN TO_CHAR(CURRENT_TIMESTAMP + ut.recurrence_interval * INTERVAL '1 week', 'YYYY-MM-DD')
        WHEN ut.recurrence_pattern = 'monthly' THEN TO_CHAR(CURRENT_TIMESTAMP + ut.recurrence_interval * INTERVAL '1 month', 'YYYY-MM-DD')
        WHEN ut.recurrence_pattern = 'yearly' THEN TO_CHAR(CURRENT_TIMESTAMP + ut.recurrence_interval * INTERVAL '1 year', 'YYYY-MM-DD')
        ELSE NULL
    END AS next_start_date,
    CASE
        WHEN ut.recurrence_pattern = 'daily' THEN TO_CHAR(CURRENT_TIMESTAMP + ut.recurrence_interval * INTERVAL '1 day', 'YYYY-MM-DD')
        WHEN ut.recurrence_pattern = 'weekly' THEN TO_CHAR(CURRENT_TIMESTAMP + ut.recurrence_interval * INTERVAL '1 week', 'YYYY-MM-DD')
        WHEN ut.recurrence_pattern = 'monthly' THEN TO_CHAR(CURRENT_TIMESTAMP + ut.recurrence_interval * INTERVAL '1 month', 'YYYY-MM-DD')
        WHEN ut.recurrence_pattern = 'yearly' THEN TO_CHAR(CURRENT_TIMESTAMP + ut.recurrence_interval * INTERVAL '1 year', 'YYYY-MM-DD')
        ELSE NULL
    END AS next_due_date,
    ut.recurrence_pattern,
    ut.recurrence_interval
FROM updated_task ut
WHERE ut.completed = TRUE 
AND ut.recurrence_pattern IS NOT NULL 
AND ut.recurrence_interval IS NOT NULL;
	`

	_, err := db.Exec(query, taskID)

	if err != nil {
		return fmt.Errorf("failed to execute task update and recurrence handling: %v", err)
	}

	return nil
}
