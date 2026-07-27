package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/harmost/agent/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	pairCmd.Flags().Bool("insecure", false, "dial the hub's gRPC endpoint without TLS (local dev only)")
	rootCmd.AddCommand(pairCmd)
}

var pairCmd = &cobra.Command{
	Use:   "pair <hub-url>",
	Short: "Pair this agent with a hub using device flow",
	Args:  cobra.ExactArgs(1),
	RunE:  runPair,
}

type deviceAuthorizeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	GRPCAddr        string `json:"grpc_addr"`
}

type deviceTokenResponse struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
}

func runPair(cmd *cobra.Command, args []string) error {
	hubURL := args[0]

	// 1. Initiate device flow.
	resp, err := http.Post(hubURL+"/api/v1/device/authorize", "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to contact hub: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var auth deviceAuthorizeResponse
	if err := json.Unmarshal(body, &auth); err != nil {
		return fmt.Errorf("failed to parse device authorize response: %w", err)
	}

	// 2. Prompt user.
	fmt.Printf("\n  Go to the following URL to approve this agent:\n\n")
	fmt.Printf("    %s\n\n", auth.VerificationURI)
	fmt.Printf("  Waiting for approval (expires in %ds)...\n\n", auth.ExpiresIn)

	// 3. Poll until approved or expired.
	interval := time.Duration(auth.Interval) * time.Second
	if interval < 5*time.Second {
		interval = 5 * time.Second
	}
	deadline := time.Now().Add(time.Duration(auth.ExpiresIn) * time.Second)

	for time.Now().Before(deadline) {
		time.Sleep(interval)

		payload, _ := json.Marshal(map[string]string{"device_code": auth.DeviceCode})
		r, err := http.Post(hubURL+"/api/v1/device/token", "application/json", bytes.NewReader(payload))
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()

		if r.StatusCode == http.StatusAccepted {
			fmt.Print(".")
			continue
		}

		var tok deviceTokenResponse
		if err := json.Unmarshal(body, &tok); err != nil {
			continue
		}

		if r.StatusCode == http.StatusOK && tok.AccessToken != "" {
			fmt.Printf("\n  Paired successfully!\n")
			insecure, _ := cmd.Flags().GetBool("insecure")
			return config.Save(&config.Config{
				HubAddr:  hubURL,
				GRPCAddr: auth.GRPCAddr,
				Token:    tok.AccessToken,
				Insecure: insecure,
			})
		}

		fmt.Printf("\n  Error: %s\n", tok.Error)
		return fmt.Errorf("pairing failed: %s", tok.Error)
	}

	return fmt.Errorf("pairing timed out — please try again")
}
