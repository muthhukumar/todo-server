package query

import (
	"log"
	"strings"
	"testing"
	"time"
)

func TestGetTastsQueryDefault(t *testing.T) {
	today := time.Now().Format("2006-01-02")

	tests := []struct {
		Inputs []string
		Query  string
		Args   []interface{}
	}{
		{Inputs: []string{"", "", ""}, Query: "select * from tasks order by created_at desc", Args: []interface{}{}},
		{Inputs: []string{"", "search", ""}, Query: "select * from tasks where name ilike '%' || $1 || '%' order by created_at desc", Args: []interface{}{"search"}},

		{Inputs: []string{"", "", "true"}, Query: "select * from tasks order by created_at desc", Args: []interface{}{}},
		{Inputs: []string{"", "", "false"}, Query: "select * from tasks where completed = false  order by created_at desc", Args: []interface{}{}},
		{Inputs: []string{"", "search", "true"}, Query: "select * from tasks where name ilike '%' || $1 || '%' order by created_at desc", Args: []interface{}{"search"}},
		{Inputs: []string{"", "search", "false"}, Query: "select * from tasks where completed = false  and name ilike '%' || $1 || '%' order by created_at desc", Args: []interface{}{"search"}},

		{Inputs: []string{"my-day", "", "true"}, Query: "select * from tasks where ((marked_today != '' and date(marked_today) = $1) or (due_date != '' and date(due_date) = $1))  order by created_at desc", Args: []interface{}{today}},
		{Inputs: []string{"my-day", "", ""}, Query: "select * from tasks where ((marked_today != '' and date(marked_today) = $1) or (due_date != '' and date(due_date) = $1))  order by created_at desc", Args: []interface{}{today}},
		{Inputs: []string{"my-day", "", "false"}, Query: "select * from tasks where ((marked_today != '' and date(marked_today) = $1) or (due_date != '' and date(due_date) = $1))  and completed = false  order by created_at desc", Args: []interface{}{today}},
		{Inputs: []string{"my-day", "search", "true"}, Query: "select * from tasks where ((marked_today != '' and date(marked_today) = $1) or (due_date != '' and date(due_date) = $1))  and name ilike '%' || $2 || '%' order by created_at desc", Args: []interface{}{today, "search"}},
		{Inputs: []string{"my-day", "search", ""}, Query: "select * from tasks where ((marked_today != '' and date(marked_today) = $1) or (due_date != '' and date(due_date) = $1))  and name ilike '%' || $2 || '%' order by created_at desc", Args: []interface{}{today, "search"}},
		{Inputs: []string{"my-day", "search", "false"}, Query: "select * from tasks where ((marked_today != '' and date(marked_today) = $1) or (due_date != '' and date(due_date) = $1))  and completed = false  and name ilike '%' || $2 || '%' order by created_at desc", Args: []interface{}{today, "search"}},

		{Inputs: []string{"important", "", ""}, Query: "select * from tasks where is_important = true order by created_at desc", Args: []interface{}{}},
		{Inputs: []string{"important", "", "false"}, Query: "select * from tasks where is_important = true and completed = false  order by created_at desc", Args: []interface{}{}},
		{Inputs: []string{"important", "", "true"}, Query: "select * from tasks where is_important = true order by created_at desc", Args: []interface{}{}},
		{Inputs: []string{"important", "search", ""}, Query: "select * from tasks where is_important = true and name ilike '%' || $1 || '%' order by created_at desc", Args: []interface{}{"search"}},
		{Inputs: []string{"important", "search", "false"}, Query: "select * from tasks where is_important = true and completed = false  and name ilike '%' || $1 || '%' order by created_at desc", Args: []interface{}{"search"}},
		{Inputs: []string{"important", "search", "true"}, Query: "select * from tasks where is_important = true and name ilike '%' || $1 || '%' order by created_at desc", Args: []interface{}{"search"}},
	}

	for _, curr := range tests {
		result, args := GetTasksQuery(curr.Inputs[0], curr.Inputs[1], curr.Inputs[2], 0, 0, "")

		for i := range args {
			if args[i] != curr.Args[i] {
				log.Fatal(curr.Query, args)
			}
		}

		if strings.ToLower(result) != curr.Query {
			log.Fatal(curr.Query, args)
		}
	}

}
