package objectstore

import "context"
import "io"

type Client struct{}
func NewClient() *Client { return &Client{} }
func (c *Client) Upload(ctx context.Context, bucket, key string, r io.Reader, size int64) error { return nil }
func (c *Client) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) { return nil, nil }
