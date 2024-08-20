package internal

import "strconv"

func ParseSize(str string) int {
	size, err := strconv.Atoi(str)

	if err != nil {
		return 0
	}

	return size
}
