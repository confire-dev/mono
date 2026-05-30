package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/confire-dev/confire/auth"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Connect your Confire account via browser (GitHub OAuth or magic link)",
	Long: `Opens your browser to confire.dev/auth, logs you in, and stores the
API key securely in the OS keychain. Zero copy-paste.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLogin()
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials from the OS keychain",
	RunE: func(cmd *cobra.Command, args []string) error {
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

// ── Login (PKCE OAuth flow) ────────────────────────────────────────────────

func runLogin() error {
	// 1. Find a free port for the local callback server
	port, err := freePort()
	if err != nil {
		return fmt.Errorf("no free port: %w", err)
	}
	callbackURL := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	// 2. Generate PKCE pair
	pkce, err := auth.NewPKCE()
	if err != nil {
		return fmt.Errorf("pkce: %w", err)
	}

	// 3. Build Supabase authorization URL
	supabaseURL := supabaseProjectURL()
	authURL := fmt.Sprintf(
		"%s/auth/v1/authorize?provider=github&response_type=code&code_challenge=%s&code_challenge_method=S256&redirect_to=%s",
		supabaseURL, pkce.Challenge, url.QueryEscape(callbackURL),
	)

	// 4. Start local callback server
	codeCh := make(chan string, 1)
	errCh  := make(chan error, 1)
	srv := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", port)}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errMsg := r.URL.Query().Get("error_description")
			errCh <- fmt.Errorf("auth failed: %s", errMsg)
			w.Write([]byte(successPage("Authentication failed — please try again.")))
			return
		}
		codeCh <- code
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(successPage("✓ Confire connected! You can close this tab.")))
	})
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// 5. Open browser
	fmt.Printf("Opening browser for login...\n")
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("\nCouldn't open browser automatically. Visit:\n%s\n\n", authURL)
	}

	// 6. Wait for callback (5-minute timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	defer func() {
		shutdownCtx, c := context.WithTimeout(context.Background(), 2*time.Second)
		defer c()
		srv.Shutdown(shutdownCtx)
	}()

	var code string
	select {
	case code = <-codeCh:
	case err = <-errCh:
		return err
	case <-ctx.Done():
		return fmt.Errorf("login timed out after 5 minutes")
	}

	// 7. Exchange code for Supabase access_token via PKCE
	accessToken, err := exchangeCode(supabaseURL, code, pkce.Verifier, callbackURL)
	if err != nil {
		return fmt.Errorf("token exchange: %w", err)
	}

	// 8. Call Worker to generate a long-lived confire API key
	apiKey, email, msg, err := generateConfireKey(accessToken)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	// 9. Store API key + email in OS keychain
	if err := auth.StoreKey(apiKey); err != nil {
		return fmt.Errorf("keychain store: %w", err)
	}
	if err := auth.StoreEmail(email); err != nil {
		_ = err // non-fatal
	}

	fmt.Printf("\n%s\n\n", msg)
	return nil
}

// exchangeCode exchanges a PKCE code for a Supabase access_token.
func exchangeCode(supabaseURL, code, verifier, redirectURI string) (string, error) {
	body := url.Values{
		"grant_type":    {"pkce"},
		"code":          {code},
		"code_verifier": {verifier},
	}
	resp, err := http.PostForm(supabaseURL+"/auth/v1/token", body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf("%s", result.Error)
	}
	return result.AccessToken, nil
}

// generateConfireKey calls the Worker /api/keys/generate with the Supabase JWT.
func generateConfireKey(supabaseJWT string) (apiKey, email, message string, err error) {
	workerURL := workerURLEnv()
	req, _ := http.NewRequest("POST", workerURL+"/api/keys/generate", nil)
	req.Header.Set("Authorization", "Bearer "+supabaseJWT)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	var result struct {
		APIKey  string `json:"apiKey"`
		Message string `json:"message"`
		User    struct {
			Email string `json:"email"`
		} `json:"user"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", "", err
	}
	if result.Error != "" {
		return "", "", "", fmt.Errorf("%s", result.Error)
	}
	return result.APIKey, result.User.Email, result.Message, nil
}

// ── whoami ─────────────────────────────────────────────────────────────────

func runWhoami() error {
	key, err := auth.LoadKey()
	if err != nil || key == "" {
		fmt.Println("Not logged in. Run: confire login")
		return nil
	}

	workerURL := workerURLEnv()
	req, _ := http.NewRequest("GET", workerURL+"/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+key)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// Offline — show cached email
		email, _ := auth.LoadEmail()
		if email != "" {
			fmt.Printf("Logged in as %s (offline)\n", email)
		} else {
			fmt.Println("Logged in (offline — can't reach Worker to show details)")
		}
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		Email string `json:"email"`
		Plan  string `json:"plan"`
		Used  int    `json:"used"`
		Limit int    `json:"limit"`
		Error string `json:"error"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Error != "" {
		fmt.Println("Not logged in. Run: confire login")
		return nil
	}

	fmt.Printf("✓ %s  ·  Plan: %s  ·  %d/%d requests used this month\n",
		result.Email, result.Plan, result.Used, result.Limit)

	// Warn at 80%
	if result.Limit > 0 && result.Used >= result.Limit*80/100 {
		fmt.Printf("  ⚠️  Upgrade for unlimited: confire.dev/upgrade\n")
	}
	return nil
}

// ── Helpers ────────────────────────────────────────────────────────────────

func freePort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	// Use a random port in range to avoid race condition
	_ = rand.New(rand.NewSource(int64(port)))
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

func supabaseProjectURL() string {
	if u := os.Getenv("SUPABASE_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}
	// Placeholder — set via SUPABASE_URL env var or confire config
	return "https://YOUR_PROJECT.supabase.co"
}

func successPage(msg string) string {
	return `<!DOCTYPE html><html><body style="font-family:monospace;padding:40px;max-width:500px">
<h2>` + msg + `</h2><p>Return to your terminal to continue.</p></body></html>`
}
