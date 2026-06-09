package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/confire-dev/confire/auth"
	"github.com/confire-dev/confire/config"
	"github.com/confire-dev/confire/hosts"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show what Confire is doing right now",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func tick(ok bool) string {
	if ok {
		return fmt.Sprintf("%s✓%s", green, reset)
	}
	return fmt.Sprintf("%s✗%s", red, reset)
}

func runStatus() error {
	fmt.Printf("\n%s[confire status]%s\n\n", bold, reset)

	cfg := config.Load()

	// ── Agents ────────────────────────────────────────────────────────────
	fmt.Printf("  %sAgents%s\n", dim, reset)
	for _, h := range hosts.Registry() {
		label := h.Label()
		if h.ComingSoon() {
			detected := h.Detect()
			if detected {
				fmt.Printf("  %s○%s  %-18s %sfound ✓  coming soon%s\n", gray, reset, label, dim, reset)
			} else {
				fmt.Printf("  %s○%s  %-18s %scoming soon%s\n", gray, reset, label, dim, reset)
			}
			continue
		}
		detected := h.Detect()
		installed := detected && h.IsInstalled(h.Preferred())
		if !detected {
			fmt.Printf("  %s○%s  %-18s %snot detected%s\n", gray, reset, label, dim, reset)
			continue
		}
		if installed {
			fmt.Printf("  %s  %-18s %sfound ✓  hook installed ✓%s\n", tick(true), label, green, reset)
		} else {
			fmt.Printf("  %s  %-18s %sfound ✓  %srun `confire setup` to install hook%s\n",
				tick(false), label, green, dim, reset)
		}
	}

	// ── Account ───────────────────────────────────────────────────────────
	fmt.Printf("\n  %sAccount%s\n", dim, reset)
	apiKey, _ := auth.LoadKey()
	hasKey := apiKey != ""
	if !hasKey {
		fmt.Printf("  %s○%s  %-18s %slocal-only mode%s\n",
			gray, reset, "Not signed in", dim, reset)
		fmt.Printf("\n  %sDashboard%s\n", dim, reset)
		fmt.Printf("  %s○%s  %-18s %soff · run `confire login` to enable%s\n",
			gray, reset, "Sync", dim, reset)
	} else {
		email, _ := auth.LoadEmail()
		if info, err := fetchAccountInfo(apiKey); err == nil {
			limit := info.EffectiveLimit
			if limit == 0 {
				limit = info.Limit
			}
			fmt.Printf("  %s  %-18s %s%s%s · %s%s plan%s · %s%d/%d%s credits\n",
				tick(true), "Signed in",
				green, info.Email, reset,
				bold, info.Plan, reset,
				cyan, info.Used, limit, reset)
			if info.PurchasedCredits > 0 {
				fmt.Printf("       %s+%d top-up credits available%s\n", dim, info.PurchasedCredits, reset)
			}
		} else if email != "" {
			fmt.Printf("  %s  %-18s %s%s%s %s(offline)%s\n",
				tick(true), "Signed in", green, email, reset, dim, reset)
		} else {
			fmt.Printf("  %s  %-18s %slogged in%s %s(key present, offline)%s\n",
				tick(true), "Signed in", green, reset, dim, reset)
		}
	}

	// ── Firewall ──────────────────────────────────────────────────────────
	fmt.Printf("\n  %sFirewall%s\n", dim, reset)
	daemonRunning := isDaemonRunning()
	firewallEnabled := cfg.IsFirewallEnabled()
	mode := cfg.EffectiveMode()
	firewallActive := daemonRunning && firewallEnabled
	if daemonRunning {
		fmt.Printf("  %s  %-18s %s%s%s\n", tick(true), "Mode", green, mode, reset)
		fmt.Printf("  %s  %-18s %srunning%s\n", tick(true), "Daemon", green, reset)
		fwColor := green
		fwLabel := "active"
		if !firewallActive {
			fwColor = dim
			fwLabel = "disabled — run `confire on`"
		}
		fmt.Printf("  %s  %-18s %s%s%s\n", tick(firewallActive), "Tool Firewall", fwColor, fwLabel, reset)
		fmt.Printf("  %s  %-18s %s%s%s\n", tick(firewallActive), "Result Firewall", fwColor, fwLabel, reset)
	} else {
		fmt.Printf("  %s  %-18s %sstopped — run `confire start`%s\n", tick(false), "Daemon", dim, reset)
		fmt.Printf("  %s○%s  %-18s %s%s (inactive)%s\n", gray, reset, "Mode", dim, mode, reset)
		fmt.Printf("  %s○%s  %-18s %sinactive%s\n", gray, reset, "Tool Firewall", dim, reset)
		fmt.Printf("  %s○%s  %-18s %sinactive%s\n", gray, reset, "Result Firewall", dim, reset)
	}

	fmt.Println()
	return nil
}

// isDaemonRunning checks whether the daemon unix socket is accepting connections.
func isDaemonRunning() bool {
	conn, err := net.DialTimeout("unix", daemonSocketPath(), 100*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// fetchAccountInfo calls Worker /api/me to get live account state.
type accountInfo struct {
	Email            string `json:"email"`
	Plan             string `json:"plan"`
	Used             int    `json:"used"`
	Limit            int    `json:"limit"`
	PurchasedCredits int    `json:"purchasedCredits"`
	EffectiveLimit   int    `json:"effectiveLimit"`
}

func fetchAccountInfo(apiKey string) (*accountInfo, error) {
	req, _ := http.NewRequest("GET", workerURLEnv()+"/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	var info accountInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	if info.Email == "" {
		return nil, fmt.Errorf("empty response")
	}
	return &info, nil
}

// workerOnline does a quick health check against the Worker.
func workerOnline() bool {
	resp, err := http.Get(workerURLEnv() + "/health")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// confireDir re-used from paths.go for the socket path.
var _ = os.DevNull // suppress unused import warning
