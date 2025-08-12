package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Banjirome/yandex-music-go/client"
)

// Example simple usage.
func main() {
	ctx := context.Background()
	cli := client.New(client.WithToken(os.Getenv("YANDEX_MUSIC_TOKEN")))
	resp, err := cli.Search.Tracks(ctx, "Muse Uprising", 0, 5)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	if resp.Result.Tracks != nil {
		for _, tr := range resp.Result.Tracks.Results {
			fmt.Println(tr.Title)
		}
	}
}
