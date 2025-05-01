package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const (
	AuthFile = "auth.json"
)

// AuthToken represents the structure of the authentication token
type AuthToken struct {
	Token string `json:"token"`
}

func main() {
	// Define CLI commands
	helpCmd := flag.NewFlagSet("help", flag.ExitOnError)
	cronCmd := flag.NewFlagSet("cronjob", flag.ExitOnError)
	authCmd := flag.NewFlagSet("auth", flag.ExitOnError)

	// Parse provided CLI arguments
	if len(os.Args) < 2 {
		fmt.Println("Expected 'help', 'cronjob' or 'auth' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "help":
		handleHelp(helpCmd)
	case "cronjob":
		handleCronJob(cronCmd)
	case "auth":
		handleAuth(authCmd)
	default:
		fmt.Println("Unknown command. Use 'help' for usage.")
		os.Exit(1)
	}
}

// handleHelp outputs a help page for all commands
func handleHelp(helpCmd *flag.FlagSet) {
	helpCmd.Parse(os.Args[2:])
	fmt.Println("Available Commands:")
	fmt.Println("  help     - Displays this help page")
	fmt.Println("  cronjob  - Executes multiple binaries at scheduled intervals")
	fmt.Println("  auth     - Authenticates with PocketBase.io and stores the token")
}

// handleCronJob executes multiple binaries like bin/cpu.go
func handleCronJob(cronCmd *flag.FlagSet) {
	cronCmd.Parse(os.Args[2:])

	binaries := []string{"bin/cpu"} // List of binaries to execute

	for _, binary := range binaries {
		go func(bin string) {
			for {
				fmt.Printf("Executing binary: %s\n", bin)
				cmd := exec.Command(bin)
				output, err := cmd.CombinedOutput()
				if err != nil {
					log.Printf("Error executing %s: %s", bin, err)
				}
				fmt.Printf("Output from %s: %s\n", bin, output)
				time.Sleep(10 * time.Second) // Sleep before the next execution
			}
		}(binary)
	}

	// Keep the main function running
	select {}
}

// handleAuth fetches a token from PocketBase.io and stores it in auth.json
func handleAuth(authCmd *flag.FlagSet) {
	authCmd.Parse(os.Args[2:])

	fmt.Print("Enter your PocketBase.io username: ")
	var username string
	fmt.Scanln(&username)

	fmt.Print("Enter your PocketBase.io password: ")
	var password string
	fmt.Scanln(&password)

	token, err := authenticateWithPocketBase(username, password)
	if err != nil {
		log.Fatalf("Authentication failed: %s", err)
	}

	authData := AuthToken{Token: token}
	saveAuthToken(authData)
	fmt.Println("Authentication successful! Token saved to auth.json.")
}

// authenticateWithPocketBase handles authentication with PocketBase.io
func authenticateWithPocketBase(username, password string) (string, error) {
	url := "https://your-pocketbase-url/api/auth" // Replace with your PocketBase.io API URL
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to authenticate, status code: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	token, ok := result["token"].(string)
	if !ok {
		return "", fmt.Errorf("failed to retrieve token from response")
	}

	return token, nil
}

// saveAuthToken stores the token in auth.json file
func saveAuthToken(authData AuthToken) {
	data, err := json.Marshal(authData)
	if err != nil {
		log.Fatalf("Failed to marshal auth token: %s", err)
	}

	if err := ioutil.WriteFile(AuthFile, data, 0644); err != nil {
		log.Fatalf("Failed to write auth token to file: %s", err)
	}
}