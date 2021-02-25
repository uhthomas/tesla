package tesla

import (
	"context"
	"fmt"
	"net/http"
)

type Option func(ctx context.Context, s *Service) error

func Client(c *http.Client) Option {
	return func(ctx context.Context, s *Service) error {
		s.c = c
		return nil
	}
}

func OAuth2(configPath, tokenPath string) Option {
	return func(ctx context.Context, s *Service) error {
		c, err := newOAuth2Client(ctx, configPath, tokenPath)
		if err != nil {
			return fmt.Errorf("new oauth2 client: %w", err)
		}
		s.c = c
		return nil
	}
}
