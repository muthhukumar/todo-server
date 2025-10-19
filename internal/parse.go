package internal

import (
	"fmt"
	"regexp"
	"strconv"
)

func SafeParseSize(str string) int {
	size, err := strconv.Atoi(str)

	if err != nil {
		return 0
	}

	return size
}

func ParseSize(str string) *int {
	if str == "" {
		return nil
	}

	size, err := strconv.Atoi(str)

	if err != nil {
		return nil
	}

	return &size
}

func ExtractTitle(input string) (string, error) {
	re := regexp.MustCompile(`<title>(.*?)</title>`)

	matches := re.FindStringSubmatch(input)

	if len(matches) > 1 {
		return matches[1], nil
	}

	return "", fmt.Errorf("no title found")
}
