package query

import (
	"fmt"
	"strings"
	"time"
)

func GetTasksQuery(filter string, searchTerm string, showCompleted string, size int, listID *int, showAllTasks string, profileId *int) (string, []interface{}) {
	var query string
	var args []interface{} = []interface{}{}
	var completedFilter string

	switch showCompleted {
	case "true":
		completedFilter = " " // No filter for completed tasks
	case "false":
		completedFilter = " t.completed = false "
	default:
		completedFilter = " "
	}

	selectClause := `
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
			t.list_id AS list_id, 
			t.profile_id AS profile_id,
			COALESCE(t.recurrence_pattern::TEXT, '') AS recurrence_pattern, 
			COALESCE(COUNT(CASE WHEN st.completed = false THEN 1 END), 0) AS incomplete_subtask_count,
			COALESCE(COUNT(st.id), 0) AS subtask_count
		FROM 
			tasks t
		LEFT JOIN 
			sub_tasks st ON st.task_id = t.id
	`

	switch filter {
	case "":
		query = selectClause
	case "my-day":
		today := time.Now().Format("2006-01-02")
		query = selectClause + `
		WHERE ((marked_today != '' AND DATE(marked_today) = $1) OR (due_date != '' AND DATE(due_date) = $1)) `
		args = append(args, today)
	case "important":
		query = selectClause + `
		WHERE t.is_important = true
		`
	default:
		query = selectClause
	}

	if profileId != nil {
		if strings.Contains(query, "WHERE") {
			query += " AND"
		} else {
			query += " WHERE"
		}
		query += fmt.Sprintf(" t.profile_id = $%d", len(args)+1)
		args = append(args, profileId)
	}

	if showCompleted == "false" {
		if strings.Contains(query, "WHERE") {
			query += " AND"
		} else {
			query += " WHERE"
		}
		query += completedFilter
	}

	if searchTerm != "" {
		if strings.Contains(query, "WHERE") {
			query += " AND"
		} else {
			query += " WHERE"
		}
		query += fmt.Sprintf(" t.name ILIKE '%%' || $%d || '%%'", len(args)+1)
		args = append(args, searchTerm)
	}

	if listID == nil {
		if filter != "important" && filter != "my-day" && showAllTasks != "true" {
			if strings.Contains(query, "WHERE") {
				query += " AND"
			} else {
				query += " WHERE"
			}
			query += " t.list_id IS NULL "
		}
	} else {
		if strings.Contains(query, "WHERE") {
			query += " AND"
		} else {
			query += " WHERE"
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
