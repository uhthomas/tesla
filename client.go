package tesla

import "net/http"

type Client struct {
	c *http.Client
}

func New(token string, opts ...Option) (*Client, error) {
	var c Client
	for _, opt := range opts {
		if err := opt(&c); err != nil {
			return nil, err
		}
	}
	return &c, nil
}
