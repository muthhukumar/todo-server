package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

type NotionBaseResponseResultPropertiesNameTitle struct {
	PlainText string `json:"plain_text"`
}

type NotionBaseResponseResultPropertiesName struct {
	Title []NotionBaseResponseResultPropertiesNameTitle `json:"title"`
}

type NotionBaseResponseResultProperties struct {
	Name NotionBaseResponseResultPropertiesName `json:"Name"`
}

type NotionBaseResponseResult struct {
	Properties NotionBaseResponseResultProperties `json:"properties"`
}

type NotionBaseResponse struct {
	Results    []NotionBaseResponseResult `json:"results"`
	NextCursor string                     `json:"next_cursor"`
	HasMore    bool                       `json:"has_more"`
}

func GetQuotesFromNotion() ([]string, error) {
	notionSecretToken := os.Getenv("NOTION_SECRET_TOKEN")
	notionDatabase := os.Getenv("NOTION_DATABASE")

	url := fmt.Sprintf("https://api.notion.com/v1/databases/%s/query", notionDatabase)

	client := &http.Client{}

	req, err := http.NewRequest("POST", url, nil)

	if err != nil {
		return nil, err
	}

	bearerToken := fmt.Sprintf("Bearer %s", notionSecretToken)

	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Authorization", bearerToken)
	req.Header.Add("Notion-Version", "2022-06-28")

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)

		if err != nil {
			return nil, errors.New(string(bodyBytes))
		}
	}

	notionResponse := &NotionBaseResponse{}

	decodeErr := json.NewDecoder(resp.Body).Decode(notionResponse)

	if decodeErr != nil {
		return nil, decodeErr
	}

	var quotes []string

	for _, result := range notionResponse.Results {
		for _, title := range result.Properties.Name.Title {
			quotes = append(quotes, title.PlainText)
		}
	}

	return quotes, nil
}
