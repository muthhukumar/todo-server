package data

import (
	"fmt"
	"testing"
)

func TestGetRandomQuotes(t *testing.T) {
	if len(GetRandomQuotes()) != MAX_QUOTES {
		t.Fatal(fmt.Sprintf("Get Random Quotes should return %v quotes", MAX_QUOTES))
	}
}
