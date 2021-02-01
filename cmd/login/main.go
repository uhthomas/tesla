package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"

	"golang.org/x/oauth2"
)

func state() string {
	var b [9]byte
	if _, err := io.ReadFull(rand.Reader, b[:]); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b[:])
}

func pxce() (challenge, sum string, err error) {
	var p [86]byte
	if _, err := io.ReadFull(rand.Reader, p[:]); err != nil {
		return "", "", fmt.Errorf("rand read full: %w", err)
	}
	challenge = base64.RawURLEncoding.EncodeToString(p[:])
	b := sha256.Sum256([]byte(challenge))
	sum = base64.RawURLEncoding.EncodeToString(b[:])
	return challenge, sum, nil
}

func Main(ctx context.Context) error {
	challenge, sum, err := pxce()
	if err != nil {
		return fmt.Errorf("pxce: %w", err)
	}

	c := &oauth2.Config{
		ClientID:     "ownerapi",
		ClientSecret: "",
		RedirectURL:  "https://auth.tesla.com/void/callback",
		Scopes:       []string{"openid email offline_access"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://auth.tesla.com/oauth2/v3/authorize",
			TokenURL: "https://auth.tesla.com/oauth2/v3/token",
		},
	}

	// go http.ListenAndServe(":9001", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	code := r.FormValue("code")
	// 	fmt.Fprintf(w, "Received code: %v\r\nYou can now safely close this browser window.", code)
	// }))

	fmt.Printf("Go to:\n\t%s\nEnter your code: ", c.AuthCodeURL(state(), oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", sum),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	))

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return err
	}

	t, err := c.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", challenge),
	)
	if err != nil {
		return err
	}

	fmt.Printf("\n%v\n", t)

	return nil
}

func main() {
	if err := Main(context.Background()); err != nil {
		log.Fatal(err)
	}
}
