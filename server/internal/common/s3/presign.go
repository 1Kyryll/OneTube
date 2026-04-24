package s3

import (
	"context"
	"net/url"
	"time"
)

func (c *Client) PresignPut(ctx context.Context, key string, contentType string, ttl time.Duration) (string, error) {
	reqParams := url.Values{}

	if contentType != "" {
		reqParams.Set("Content-Type", contentType)
	}

	u, err := c.Public.PresignedPutObject(ctx, c.Bucket, key, ttl)
	if err != nil {
		return "", err
	}

	return u.String(), nil
}

func (c *Client) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	u, err := c.Public.PresignedGetObject(ctx, c.Bucket, key, ttl, url.Values{})
	if err != nil {
		return "", err
	}

	return u.String(), nil
}
