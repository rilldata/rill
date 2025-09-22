package devtool

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func GithubTokensCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gh-token",
		Short: "Initiates GitHub device flow to generate personal access tokens",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := godotenv.Load(".env")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading .env: %v\n", err)
				os.Exit(1)
			}
			// 1) Get device code
			clientID := os.Getenv("RILL_ADMIN_GITHUB_CLIENT_ID")
			if clientID == "" {
				return fmt.Errorf("run `rill devtoot dotenv refresh` to generate .env with GitHub App credentials")
			}
			payload := url.Values{
				"client_id": {clientID},
			}

			resp, err := post("https://github.com/login/device/code", payload)
			if err != nil {
				return fmt.Errorf("error getting device code: %w", err)
			}

			verificationURI := resp["verification_uri"].(string)
			userCode := resp["user_code"].(string)
			deviceCode := resp["device_code"].(string)

			fmt.Fprintf(os.Stderr, "Open %s and enter code: %s\n", verificationURI, userCode)

			// 2) Poll for token
			output := map[string]string{}
			for {
				time.Sleep(5 * time.Second)
				fmt.Printf("Polling for token...\n")
				tokenResp, err := post(
					"https://github.com/login/oauth/access_token",
					url.Values{
						"client_id":   {clientID},
						"device_code": {deviceCode},
						"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
					},
				)
				if err != nil {
					return fmt.Errorf("Error polling for token: %w", err)
				}

				if errorStr, ok := tokenResp["error"].(string); ok {
					if errorStr == "authorization_pending" || errorStr == "slow_down" {
						continue
					}

					errorDesc := ""
					if desc, ok := tokenResp["error_description"]; ok {
						errorDesc = desc.(string)
					}
					return fmt.Errorf("error: %s (%s)", errorStr, errorDesc)
				}

				// For GitHub App user tokens: access_token=ghu_..., refresh_token=ghr_...
				output["access_token"] = tokenResp["access_token"].(string)
				output["refresh_token"] = tokenResp["refresh_token"].(string)
				break
			}

			// set token in gh.env file
			f, err := os.OpenFile(".github_env", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
			if err != nil {
				return fmt.Errorf("error creating gh.env: %w", err)
			}
			defer f.Close()

			for k, v := range output {
				_, err := f.WriteString(fmt.Sprintf("GH_%s=%s\n", strings.ToUpper(k), v))
				if err != nil {
					return fmt.Errorf("error writing to gh.env: %w", err)
				}
			}
			fmt.Printf("Tokens written to .github_env file\n")
			return nil
		},
	}

	return cmd
}

func post(url string, data url.Values) (map[string]interface{}, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}
