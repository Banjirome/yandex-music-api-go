package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Downloader упрощённый аналог DataDownloader из C#.
type Downloader struct{ c *Client }

// Stream возвращает поток ответа.
func (d *Downloader) Stream(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		resp.Body.Close()
		return nil, fmt.Errorf("download failed: %s (%d) %s", resp.Status, resp.StatusCode, string(b))
	}
	return resp.Body, nil
}

// Bytes загружает контент как байты.
func (d *Downloader) Bytes(ctx context.Context, url string) ([]byte, error) {
	rc, err := d.Stream(ctx, url)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

// ToFile сохраняет ресурс в файл.
func (d *Downloader) ToFile(ctx context.Context, url, filename string) error {
	rc, err := d.Stream(ctx, url)
	if err != nil {
		return err
	}
	defer rc.Close()
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, rc)
	return err
}
