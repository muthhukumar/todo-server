package utils

import (
	"fmt"
	"log"
)

func Assert(condition bool, message string) {
	if !condition {
		log.Fatalf("Assertion failed: %s", message)
	}

	fmt.Println("Assert pass: v", message)
}
