package internal

import (
	"fmt"
	"io"
	"strconv"

	"golang.org/x/net/html"
)

func ParseSize(str string) int {
	size, err := strconv.Atoi(str)

	if err != nil {
		return 0
	}

	return size
}

func ParseHTMLTitle(r io.Reader) (string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", err
	}

	var title string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
			title = n.FirstChild.Data
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	if title == "" {
		return "", fmt.Errorf("title tag not found")
	}

	return title, nil
}
