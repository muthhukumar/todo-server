package data

import (
	"testing"
)

func TestGetRandomQuotes(t *testing.T) {
	if len(GetRandomQuotes(Quotes, 2)) != 2 {
		t.Fatalf("Get Random Quotes should return %v quotes", 2)
	}
}
