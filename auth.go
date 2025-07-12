package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/oauth2"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	// Start a local HTTP server to handle the callback
	codeCh := make(chan string)
	errCh := make(chan error)

	server := &http.Server{Addr: ":8080"}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no authorization code received")
			return
		}

		// Send success page
		fmt.Fprintf(w, `
			<html>
			<head><title>Authorization Success</title></head>
			<body style="font-family: Arial, sans-serif; text-align: center; padding: 50px;">
				<h1 style="color: green;">✅ Authorization Successful!</h1>
				<p>You can now close this window and return to the terminal.</p>
			</body>
			</html>
		`)

		codeCh <- code
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Update the redirect URI to match our local server
	config.RedirectURL = "http://localhost:8080"

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Opening browser for authorization...\n")
	fmt.Printf("If the browser doesn't open automatically, go to: %v\n", authURL)

	// Try to open the browser automatically
	openBrowser(authURL)

	var authCode string
	select {
	case authCode = <-codeCh:
		fmt.Println("✅ Authorization code received successfully!")
	case err := <-errCh:
		log.Fatalf("Error during authorization: %v", err)
	case <-time.After(5 * time.Minute):
		log.Fatalf("Authorization timed out after 5 minutes")
	}

	// Shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// openBrowser tries to open the URL in a browser
func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Printf("Could not open browser automatically: %v\n", err)
	}
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// Retrieve a token from existing file, don't start OAuth flow
func getClientFromExistingToken(config *oauth2.Config) (*http.Client, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("unable to determine executable path: %v", err)
	}
	tokenPath := filepath.Join(filepath.Dir(execPath), "token.json")
	tok, err := tokenFromFile(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load existing token from %s: %v. Please run 'agenda-mcp auth' first", tokenPath, err)
	}
	// Use the config passed as parameter to create the client with the loaded token
	return config.Client(context.Background(), tok), nil
}
