package chrome

import (
	"context"
	"encoding/base64"
	"os"
	"time"
	"todo-server/utils"

	"github.com/chromedp/cdproto/page"
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

func ScreenshotWebsite(url string) (string, []byte, error) {
	chromePath := os.Getenv("CHROME_PATH")

	utils.Assert(chromePath != "", "Chrome Path ENV value is not set")

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(chromePath),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var buf []byte

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		fullScreenshot(url, 854, 480, &buf),
	)

	if err != nil {
		return "", nil, err
	}

	imgBase64Str := base64.StdEncoding.EncodeToString(buf)

	return imgBase64Str, buf, nil

}

func fullScreenshot(urlstr string, width, height int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.EmulateViewport(int64(width), int64(height)),
		chromedp.Navigate(urlstr),
		chromedp.Sleep(2 * time.Second), // wait for the page to load
		// chromedp.FullScreenshot(res, 20),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			*res, err = page.CaptureScreenshot().
				WithClip(&page.Viewport{
					X:      0,
					Y:      0,
					Width:  float64(width),
					Height: float64(height),
					Scale:  1,
				}).
				WithQuality(1).
				Do(ctx)
			return err
		}),
	}
}
