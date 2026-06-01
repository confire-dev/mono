package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/confire-dev/confire/auth"
	"github.com/spf13/cobra"
)

const (
	topupPollInterval = 2 * time.Second
	topupPollTimeout  = 5 * time.Minute
)

var topupCmd = &cobra.Command{
	Use:   "topup [packs]",
	Short: "Buy extra optimization credits ($5 per 5,000-call pack)",
	Long: `Opens Stripe Checkout for one-time credit top-up packs.
Each pack adds 5,000 cloud optimization calls for $5.

Examples:
  confire topup       buy 1 pack
  confire topup 3     buy 3 packs (15,000 credits)`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTopup(args)
	},
}

func init() {
	rootCmd.AddCommand(topupCmd)
}

type accountCredits struct {
	Email            string `json:"email"`
	Used             int    `json:"used"`
	Limit            int    `json:"limit"`
	PurchasedCredits int    `json:"purchasedCredits"`
	EffectiveLimit   int    `json:"effectiveLimit"`
	Error            string `json:"error"`
}

func runTopup(args []string) error {
	apiKey, err := auth.LoadKey()
	if err != nil || apiKey == "" {
		return fmt.Errorf("not logged in — run `confire login` first")
	}

	quantity := 1
	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 1 || n > 20 {
			return fmt.Errorf("pack count must be 1–20")
		}
		quantity = n
	}

	baseline, err := fetchAccountCredits(apiKey)
	if err != nil {
		return fmt.Errorf("could not reach server: %w", err)
	}

	body, _ := json.Marshal(map[string]int{"quantity": quantity})
	req, err := http.NewRequest("POST", workerURLEnv()+"/api/topup/create", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("checkout request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		_ = json.Unmarshal(respBody, &errResp)
		msg := errResp.Message
		if msg == "" {
			msg = errResp.Error
		}
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return fmt.Errorf("top-up failed: %s", msg)
	}

	var checkout struct {
		URL          string `json:"url"`
		TotalCredits int    `json:"totalCredits"`
		Quantity     int    `json:"quantity"`
	}
	if err := json.Unmarshal(respBody, &checkout); err != nil {
		return fmt.Errorf("invalid checkout response: %w", err)
	}
	if checkout.URL == "" {
		return fmt.Errorf("checkout URL missing from server response")
	}

	fmt.Printf("\nOpening checkout for %d pack(s) — %s credits ($%d)...\n\n",
		checkout.Quantity, formatInt(checkout.TotalCredits), checkout.Quantity*5)
	fmt.Printf("%s\n\n", checkout.URL)
	fmt.Println("If the browser didn't open, copy the URL above.")
	_ = openBrowser(checkout.URL)

	fmt.Println("Waiting for payment confirmation...")
	deadline := time.Now().Add(topupPollTimeout)
	for time.Now().Before(deadline) {
		time.Sleep(topupPollInterval)
		current, err := fetchAccountCredits(apiKey)
		if err != nil {
			continue
		}
		if current.PurchasedCredits > baseline.PurchasedCredits {
			added := current.PurchasedCredits - baseline.PurchasedCredits
			fmt.Printf("\n%s✓ Top-up complete%s — +%s credits (%d/%d used this period)\n\n",
				green, reset,
				formatInt(added),
				current.Used, current.EffectiveLimit)
			return nil
		}
	}

	return fmt.Errorf("payment not confirmed within %v — check your email or run `confire whoami`", topupPollTimeout)
}

func fetchAccountCredits(apiKey string) (*accountCredits, error) {
	req, err := http.NewRequest("GET", workerURLEnv()+"/api/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var info accountCredits
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	if info.Error != "" {
		return nil, fmt.Errorf(info.Error)
	}
	return &info, nil
}

func formatInt(n int) string {
	return strconv.Itoa(n)
}
