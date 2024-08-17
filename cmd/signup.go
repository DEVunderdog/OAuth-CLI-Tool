package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/DEVunderdog/concept_OAuth/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var SignupCmd = &cobra.Command{
	Use:   "signup",
	Short: "User can signup via github",
	Run:   runSignup,
}

func runSignup(cmd *cobra.Command, args []string) {
	var deviceCode string
	ctx := context.Background()
	interval := 7 * time.Second
	duration := 15 * time.Minute
	polling_url := "https://github.com/login/oauth/access_token"

	url := "https://github.com/login/device/code"

	data := map[string]string{
		"client_id": viper.GetString("CLIENT_ID"),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending requests: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	if strings.HasPrefix(string(body), "device_code=") {
		parts := strings.Split(string(body), "&")
		for index, part := range parts {
			if index == 0 {
				deviceCode = strings.TrimPrefix(part, "device_code=")
			}
			if index == 3 {
				userCode := strings.TrimPrefix(part, "user_code=")
				log.Printf("\nUser Code: %v\n", userCode)
			}
		}
	} else {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Println("Error parsing JSON response:", err)
			fmt.Println("Full response body:", string(body))
		} else {
			fmt.Println("Parsed JSON response:", result)
		}
	}

	browser_provider, url_redirect, err := utils.OpenBrowser("https://github.com/login/device")
	if err != nil {
		log.Fatalf(err.Error())
	}

	browser_cmd := exec.Command(*browser_provider, *url_redirect)
	// browser_cmd.Stdout = os.Stdout
	browser_cmd.Stderr = os.Stderr
	browser_cmd.Run()

	err = githubAuthServer(ctx, polling_url, deviceCode, interval, duration)
	if err != nil {
		fmt.Printf("Polling ended: %v", err)
	}
}

func githubAuthServer(ctx context.Context, polling_url string, device_code string, interval, duration time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			if timeoutCtx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("polling duration exceeded")
			}
			return timeoutCtx.Err()
		case <-ticker.C:
			accessToken, err := makeRequest(polling_url, device_code)
			if err != nil {
				fmt.Printf("\nError making request: %v\n", err)
			} else if accessToken != nil {
				fmt.Println("Access Token received: ", *accessToken)
				return nil
			}
		}
	}
}

func makeRequest(url string, device_code string) (*string, error) {
	data := map[string]string{
		"client_id":   viper.GetString("CLIENT_ID"),
		"device_code": device_code,
		"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err == nil {
		fmt.Println("Parsed JSON Response: ", result)
		if accessToken, ok := result["access_token"].(string); ok {
			return &accessToken, nil
		}
	}

	bodyStr := string(body)
	if strings.Contains(bodyStr, "&") && strings.Contains(bodyStr, "=") {
		parts := strings.Split(bodyStr, "&")
		for _, part := range parts {
			if strings.HasPrefix(part, "access_token") {
				accessToken := strings.TrimPrefix(part, "access_token=")
				return &accessToken, nil
			}
		}
	}

	return nil, nil
}
