package cmd

import (
	"bytes"
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
	Use: "signup",
	Short: "User can signup via github",
	Run: runSignup,
}

func runSignup(cmd *cobra.Command, args []string){
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
			if index == 3 {
				fmt.Println(part)
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
	browser_cmd.Stdout = os.Stdout
	browser_cmd.Stderr = os.Stderr
	browser_cmd.Run()

}