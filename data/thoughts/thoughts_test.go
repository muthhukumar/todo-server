package data

import (
	"testing"
)

func TestGetRandomQuotes(t *testing.T) {
	if len(GetRandomQuotes(Quotes)) != MAX_QUOTES {
		t.Fatalf("Get Random Quotes should return %v quotes", MAX_QUOTES)
	}
}
