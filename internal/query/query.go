package query

import (
	"fmt"
	"time"
)

func GetTasksQuery(filter string, searchTerm string, showCompleted string, random string, size int) (string, []interface{}) {
	var query string
	var args []interface{} = []interface{}{}
	var completedFilter string

	// Determine the completed filter condition based on the showCompleted parameter
	switch showCompleted {
	case "true":
		completedFilter = " "
	case "false":
		completedFilter = " completed = false "
	default:
		completedFilter = " " // No filter for completed status
	}

	switch filter {
	case "":
		query = "SELECT * FROM tasks"
	case "my-day":
		today := time.Now().Format("2006-01-02")
		query = "SELECT * FROM tasks WHERE ((marked_today != '' AND DATE(marked_today) = $1) OR (due_date != '' AND DATE(due_date) = $1)) "
		args = append(args, today)
	case "important":
		query = "SELECT * FROM tasks where is_important = true"
	default:
		query = "SELECT * FROM tasks"
	}

	// If show completed is true then we don't have to add the filter.
	if showCompleted != "" && showCompleted == "false" {
		if len(args) > 0 || filter == "important" || filter == "my-day" {
			query += " AND"
		} else {
			query += " WHERE"
		}

		query += completedFilter
	}

	if searchTerm != "" {
		if len(args) > 0 || filter == "important" || filter == "my-day" || (showCompleted != "" && showCompleted == "false") {
			query += " AND"
		} else {
			query += " WHERE"
		}
		if filter == "my-day" {
			query += " name ILIKE '%' || $2 || '%'"
		} else {
			query += " name ILIKE '%' || $1 || '%'"
		}

		args = append(args, searchTerm)
	}

	if random == "true" {
		query += " ORDER BY RANDOM() "
	} else {
		query += " ORDER BY created_at DESC"
	}

	if size > 0 {
		query += fmt.Sprintf(" LIMIT $%d ", len(args)+1)
		args = append(args, size)
	}

	return query, args
}
