package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/confire-dev/confire/auth"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Connect your Confire account via browser",
	Long: `Opens the Confire login page, authenticates you, and stores the
API key securely in the OS keychain. Zero copy-paste.

Use --local to authenticate against local dev servers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLogin()
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials from the OS keychain",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Revoke the key on the server first so the dashboard reflects immediately.
		if key, _ := auth.LoadKey(); key != "" {
			client := &http.Client{Timeout: 4 * time.Second}
			req, err := http.NewRequest(http.MethodPost, workerURLEnv()+"/api/keys/revoke-self", nil)
			if err == nil {
				req.Header.Set("Authorization", "Bearer "+key)
				resp, err := client.Do(req)
				if err == nil {
					resp.Body.Close()
				}
			}
		}
		if err := auth.DeleteKey(); err != nil {
			return fmt.Errorf("keychain: %w", err)
		}
		fmt.Println("✓ Logged out")
		return nil
	},
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current account and usage",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWhoami()
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(whoamiCmd)
}

// ── Login ─────────────────────────────────────────────────────────────────

func runLogin() error {
	// Already logged in — verify the key is still valid before trusting the keychain.
	if key, _ := auth.LoadKey(); key != "" {
		if keyIsValid(key) {
			email, _ := auth.LoadEmail()
			if email != "" {
				fmt.Printf("Already logged in as %s.\n", email)
			} else {
				fmt.Println("Already logged in.")
			}
			fmt.Println("Run `confire whoami` to check your account, or `confire logout` to switch.")
			return nil
		}
		// Key revoked or expired — clear it and fall through to re-login.
		fmt.Println("Your API key has been revoked. Re-authenticating...")
		_ = auth.DeleteKey()
	}

	deviceID, err := auth.DeviceID()
	if err != nil {
		deviceID = "unknown"
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	// 1. Find a free port for the local callback server
	port, err := freePort()
	if err != nil {
		return fmt.Errorf("no free port: %w", err)
	}
	callbackURL := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	// 2. Build the platform CLI login URL
	params := url.Values{
		"callback":    {callbackURL},
		"device_id":   {deviceID},
		"device_name": {hostname},
		"cli_version": {buildVersion},
	}
	loginURL := platformURL() + "/cli/login?" + params.Encode()

	// 3. Start local callback server — waits for ?api_key=...&email=...&message=...
	type result struct {
		apiKey  string
		email   string
		message string
		err     error
	}
	resultCh := make(chan result, 1)

	mux := http.NewServeMux()
	srv := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", port), Handler: mux}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if errMsg := q.Get("error"); errMsg != "" {
			resultCh <- result{err: fmt.Errorf("auth failed: %s", errMsg)}
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(callbackPage("Authorization failed — please try again.", true)))
			return
		}
		apiKey := q.Get("api_key")
		if apiKey == "" {
			resultCh <- result{err: fmt.Errorf("no API key received")}
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(callbackPage("Authorization failed — no key received.", true)))
			return
		}
		resultCh <- result{
			apiKey:  apiKey,
			email:   q.Get("email"),
			message: q.Get("message"),
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(callbackPage("✓ Confire connected! You can close this tab.", false)))
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			resultCh <- result{err: err}
		}
	}()
	defer func() {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer shutCancel()
		srv.Shutdown(shutCtx)
	}()

	// 4. Always show the URL so it can be copied to any browser, then try to open it.
	fmt.Printf("\nOpening browser for login...\n\n")
	fmt.Printf("  %s\n\n", loginURL)
	fmt.Printf("If the browser didn't open, copy the URL above.\n")
	fmt.Printf("Waiting for authorization... (Ctrl+C to cancel)\n\n")
	openBrowser(loginURL) // best-effort; URL already visible above

	// 5. Wait for the CLI callback (5-minute timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var res result
	select {
	case res = <-resultCh:
	case <-ctx.Done():
		return fmt.Errorf("login timed out after 5 minutes")
	}
	if res.err != nil {
		return res.err
	}

	// 6. Store key + email in OS keychain
	if err := auth.StoreKey(res.apiKey); err != nil {
		return fmt.Errorf("keychain store: %w", err)
	}
	if res.email != "" {
		_ = auth.StoreEmail(res.email)
	}

	go postCLIEvent("login_completed")

	msg := res.message
	if msg == "" {
		msg = fmt.Sprintf("✓ Logged in as %s", res.email)
	}
	fmt.Printf("\n%s\n\n", msg)
	return nil
}

// ── whoami ─────────────────────────────────────────────────────────────────

func runWhoami() error {
	key, err := auth.LoadKey()
	if err != nil || key == "" {
		fmt.Println("Not logged in. Run: confire login")
		return nil
	}

	req, _ := http.NewRequest("GET", workerURLEnv()+"/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+key)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		email, _ := auth.LoadEmail()
		if email != "" {
			fmt.Printf("Logged in as %s (offline)\n", email)
		} else {
			fmt.Println("Logged in (offline — can't reach server)")
		}
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		Email            string `json:"email"`
		Plan             string `json:"plan"`
		Used             int    `json:"used"`
		Limit            int    `json:"limit"`
		PurchasedCredits int    `json:"purchasedCredits"`
		EffectiveLimit   int    `json:"effectiveLimit"`
		Error            string `json:"error"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Error != "" {
		fmt.Println("Not logged in. Run: confire login")
		return nil
	}

	limit := result.EffectiveLimit
	if limit == 0 {
		limit = result.Limit
	}
	fmt.Printf("✓ %s  ·  Plan: %s  ·  %d/%d requests used this period\n",
		result.Email, result.Plan, result.Used, limit)
	if result.PurchasedCredits > 0 {
		fmt.Printf("  +%d top-up credits available\n", result.PurchasedCredits)
	}
	if result.Limit > 0 && result.Used >= result.Limit*80/100 && result.PurchasedCredits == 0 {
		fmt.Printf("  ⚠️  Need more? Run `confire topup` or upgrade at confire.dev/upgrade\n")
	}
	return nil
}

// keyIsValid does a quick server check — returns false on 401 (revoked/expired).
// Network errors are treated as valid so offline use still works.
func keyIsValid(key string) bool {
	client := &http.Client{Timeout: 4 * time.Second}
	req, err := http.NewRequest(http.MethodGet, workerURLEnv()+"/api/me", nil)
	if err != nil {
		return true
	}
	req.Header.Set("Authorization", "Bearer "+key)
	resp, err := client.Do(req)
	if err != nil {
		return true // offline — assume valid
	}
	resp.Body.Close()
	return resp.StatusCode != http.StatusUnauthorized
}

// ── Helpers ───────────────────────────────────────────────────────────────

func freePort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port, nil
}

func openBrowser(u string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd, args = "open", []string{u}
	case "linux":
		cmd, args = "xdg-open", []string{u}
	case "windows":
		cmd, args = "rundll32", []string{"url.dll,FileProtocolHandler", u}
	default:
		return fmt.Errorf("unsupported OS")
	}
	return exec.Command(cmd, args...).Start()
}

func callbackPage(msg string, isError bool) string {
	color := "#22c55e"
	if isError {
		color = "#ef4444"
	}
	return `<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8">
<style>*{box-sizing:border-box;margin:0}body{font-family:monospace;display:flex;align-items:center;justify-content:center;min-height:100vh;background:#09090b;color:#fafafa;padding:2rem}</style>
</head><body><div style="text-align:center;max-width:400px">
<p style="font-size:1.1rem;color:` + color + `">` + msg + `</p>
<p style="margin-top:1rem;color:#71717a;font-size:.85rem">Return to your terminal to continue.</p>
</div></body></html>`
}
