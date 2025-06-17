package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var sessionToken string

func LoginWithKey() error {
	url := "https://api.topstepx.com/api/Auth/loginKey"

	reqBody := map[string]string{
		"userName": os.Getenv("PROJECTX_USERNAME"),
		"apiKey":   os.Getenv("PROJECTX_API_KEY"),
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("accept", "text/plain")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result struct {
		Token        string `json:"token"`
		Success      bool   `json:"success"`
		ErrorCode    int    `json:"errorCode"`
		ErrorMessage string `json:"errorMessage"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("unmarshaling response: %w", err)
	}

	if !result.Success || result.ErrorCode != 0 {
		return fmt.Errorf("login failed: %s", result.ErrorMessage)
	}

	sessionToken = result.Token
	err = os.Setenv("PROJECTX_SESSION_TOKEN", sessionToken)
	if err != nil {
		return err
	}

	// Optionally: write token to file for the proxy
	_ = os.WriteFile("token.txt", []byte(sessionToken), 0600)

	return nil
}

func GetSessionToken() string {
	return os.Getenv("PROJECTX_SESSION_TOKEN")
}

func SearchActiveAccount() (string, error) {
	url := "https://api.topstepx.com/api/Account/search"

	reqBody := map[string]bool{
		"onlyActiveAccounts": true,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("accept", "text/plain")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+GetSessionToken())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Printf("Account search response: %s", string(body))
	var result struct {
		Accounts []struct {
			ID        int     `json:"id"`
			Name      string  `json:"name"`
			Balance   float64 `json:"balance"`
			CanTrade  bool    `json:"canTrade"`
			IsVisible bool    `json:"isVisible"`
			Simulated bool    `json:"simulated"`
		} `json:"accounts"`
		Success      bool   `json:"success"`
		ErrorCode    int    `json:"errorCode"`
		ErrorMessage string `json:"errorMessage"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("unmarshaling response: %w", err)
	}

	if !result.Success || result.ErrorCode != 0 {
		return "", fmt.Errorf("account search failed: %s", result.ErrorMessage)
	}

	for _, acct := range result.Accounts {
		if acct.CanTrade && acct.Simulated {
			if os.Getenv("BACKTEST_MODE") == "true" {
				if containsIgnoreCase(acct.Name, "practice") {
					return fmt.Sprintf("%d", acct.ID), nil
				}
			} else {
				return fmt.Sprintf("%d", acct.ID), nil
			}
		}
	}

	return "", fmt.Errorf("no tradable simulated account found")
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
