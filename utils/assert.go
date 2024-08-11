package utils

import (
	"log"
)

func Assert(condition bool, message string) {
	if !condition {
		log.Fatalf("Assertion failed: %s", message)
	}

	log.Println("Assert pass: ", message)
}
