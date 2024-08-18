package todaytasks

import (
	"os"
	"path/filepath"
	"text/template"
)

func TodayTasksEmailTemplate() (*template.Template, error) {
	dir, err := os.Getwd()

	if err != nil {
		return nil, err
	}

	htmlPath := filepath.Join(dir, "internal/templates/today-tasks/index.html")

	html, err := os.ReadFile(htmlPath)

	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("TodaysTasks").Parse(string(html))

	if err != nil {
		return nil, err
	}

	return tmpl, nil
}
