package data

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

// TODO: maybe rename this thing here later. Not sure what to rename but in the payload we have to send it as start cursor but we get the value as next_cursor
type NextCursorPayload struct {
	NextCursor string `json:"start_cursor"`
}

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

	hasMore := true
	iterations, ITERATION_LIMIT := 0, 5
	var nextCursor string

	var quotes []string

	for hasMore && iterations <= ITERATION_LIMIT {
		jsonData, err := json.Marshal(createPayload(nextCursor))

		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

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

		for _, result := range notionResponse.Results {
			for _, title := range result.Properties.Name.Title {
				quotes = append(quotes, title.PlainText)
			}
		}

		hasMore = notionResponse.HasMore
		nextCursor = notionResponse.NextCursor

		iterations += 1
	}

	return quotes, nil
}

func createPayload(nextCursor string) interface{} {
	if nextCursor == "" {
		return map[string]string{}
	}

	return NextCursorPayload{NextCursor: nextCursor}

}
