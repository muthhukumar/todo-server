package chrome

import (
	"context"
	"os"
	"time"
	"todo-server/utils"

	"github.com/chromedp/chromedp"
)

func GetTitleFromURLUsingChrome(url string) (string, error) {
	chromePath := os.Getenv("CHROME_PATH")

	utils.Assert(chromePath != "", "Chrome Path ENV value is not set")

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(chromePath),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var pageTitle string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Sleep(2*time.Second),
		chromedp.Evaluate(`document.title`, &pageTitle),
	)

	if err != nil {
		return "", err
	}

	return pageTitle, nil
}
