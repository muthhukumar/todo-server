package query

import (
	"fmt"
	"time"
)

func GetTasksQuery(filter string, searchTerm string, showCompleted string, size int, listID int, showAllTasks string) (string, []interface{}) {
	var query string
	var args []interface{} = []interface{}{}
	var completedFilter string

	// Determine the completed filter condition based on the showCompleted parameter
	switch showCompleted {
	case "true":
		completedFilter = " " // No filter for completed tasks
	case "false":
		completedFilter = " t.completed = false " // Explicitly refer to tasks table's completed column
	default:
		completedFilter = " " // No filter for completed status
	}

	switch filter {
	case "":
		query = `
		SELECT 
			t.id,
			t.name,
			t.completed,
			t.completed_on,
			t.created_at,
			t.marked_today,
			t.is_important,
			t.due_date,
			t.metadata,
			COALESCE(t.list_id, 0) AS list_id, 
			COALESCE(t.recurrence_pattern::TEXT, '') AS recurrence_pattern, 
			COALESCE(COUNT(CASE WHEN st.completed = false THEN 1 END), 0) AS incomplete_subtask_count,
			COALESCE(COUNT(st.id), 0) AS subtask_count
		FROM 
			tasks t
		LEFT JOIN 
			sub_tasks st ON st.task_id = t.id  
		`

	case "my-day":
		today := time.Now().Format("2006-01-02")
		query = `
		SELECT 
			t.id,
			t.name,
			t.completed,
			t.completed_on,
			t.created_at,
			t.marked_today,
			t.is_important,
			t.due_date,
			t.metadata,
			COALESCE(t.list_id, 0) AS list_id, 
			COALESCE(t.recurrence_pattern::TEXT, '') AS recurrence_pattern, 
			COALESCE(COUNT(CASE WHEN st.completed = false THEN 1 END), 0) AS incomplete_subtask_count,
			COALESCE(COUNT(st.id), 0) AS subtask_count
		FROM 
			tasks t
		LEFT JOIN 
			sub_tasks st ON st.task_id = t.id
		WHERE ((marked_today != '' AND DATE(marked_today) = $1) OR (due_date != '' AND DATE(due_date) = $1)) `
		args = append(args, today)
	case "important":
		query = `
		SELECT 
			t.id,
			t.name,
			t.completed,
			t.completed_on,
			t.created_at,
			t.marked_today,
			t.is_important,
			t.due_date,
			t.metadata,
			COALESCE(t.list_id, 0) AS list_id, 
			COALESCE(t.recurrence_pattern::TEXT, '') AS recurrence_pattern, 
			COALESCE(COUNT(CASE WHEN st.completed = false THEN 1 END), 0) AS incomplete_subtask_count,
			COALESCE(COUNT(st.id), 0) AS subtask_count
		FROM 
			tasks t
		LEFT JOIN 
			sub_tasks st ON st.task_id = t.id 
		WHERE t.is_important = true
		`
	default:
		query = `
		SELECT 
			t.id,
			t.name,
			t.completed,
			t.completed_on,
			t.created_at,
			t.marked_today,
			t.is_important,
			t.due_date,
			t.metadata,
			COALESCE(t.list_id, 0) AS list_id, 
			COALESCE(t.recurrence_pattern::TEXT, '') AS recurrence_pattern, 
			COALESCE(COUNT(CASE WHEN st.completed = false THEN 1 END), 0) AS incomplete_subtask_count,
			COALESCE(COUNT(st.id), 0) AS subtask_count
		FROM 
			tasks t
		LEFT JOIN 
			sub_tasks st ON st.task_id = t.id
		`
	}

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
			query += " t.name ILIKE '%' || $2 || '%'"
		} else {
			query += " t.name ILIKE '%' || $1 || '%'"
		}

		args = append(args, searchTerm)
	}

	if listID == 0 {
		if filter != "important" && filter != "my-day" && showAllTasks != "true" {
			if len(args) > 0 || (showCompleted != "" && showCompleted == "false") || searchTerm != "" {
				query += " AND "
			} else {
				query += " WHERE "
			}
			query += " t.list_id IS NULL "
		}
	} else {
		if len(args) > 0 || filter == "important" || filter == "my-day" || (showCompleted != "" && showCompleted == "false") || searchTerm != "" {
			query += " AND "
		} else {
			query += " WHERE "
		}
		query += fmt.Sprintf(" t.list_id = $%d ", len(args)+1)
		args = append(args, listID)
	}

	query += " GROUP BY t.id "
	query += " ORDER BY t.created_at DESC"

	if size > 0 {
		query += fmt.Sprintf(" LIMIT $%d ", len(args)+1)
		args = append(args, size)
	}

	return query, args
}
