package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/confire-dev/confire/receipt"
	"github.com/spf13/cobra"
)

var receiptCmd = &cobra.Command{
	Use:   "receipt",
	Short: "Inspect and verify signed provenance receipts",
}

var receiptListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recorded receipt chains",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := filepath.Join(confireDir(), "receipts")
		entries, err := os.ReadDir(dir)
		if os.IsNotExist(err) {
			fmt.Println("No receipts directory found — start the daemon to begin recording.")
			return nil
		}
		if err != nil {
			return err
		}

		var chains []os.DirEntry
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".jsonl") {
				chains = append(chains, e)
			}
		}

		if len(chains) == 0 {
			fmt.Println("No receipt chains recorded yet.")
			return nil
		}

		fmt.Printf("%-40s  %8s  %s\n", "SESSION / CHAIN", "RECEIPTS", "LAST UPDATED")
		fmt.Println(strings.Repeat("-", 70))
		for _, e := range chains {
			key := strings.TrimSuffix(e.Name(), ".jsonl")
			metaPath := filepath.Join(dir, key+".meta")
			count := "?"
			updated := ""
			if data, err := os.ReadFile(metaPath); err == nil {
				var m receipt.Meta
				if json.Unmarshal(data, &m) == nil {
					count = fmt.Sprintf("%d", m.ReceiptCount)
					if !m.UpdatedAt.IsZero() {
						updated = m.UpdatedAt.Local().Format(time.RFC822)
					}
				}
			}
			fmt.Printf("%-40s  %8s  %s\n", key, count, updated)
		}
		return nil
	},
}

var receiptVerifyCmd = &cobra.Command{
	Use:   "verify <session-id>",
	Short: "Verify the integrity of a receipt chain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionID := args[0]
		dir := filepath.Join(confireDir(), "receipts")

		var path string
		if strings.HasSuffix(sessionID, ".jsonl") {
			path = sessionID // treat as direct file path
		} else {
			path = filepath.Join(dir, sessionID+".jsonl")
		}

		receipts, err := receipt.ReadChainFromFile(path)
		if err != nil {
			return fmt.Errorf("cannot read chain %q: %w", sessionID, err)
		}
		if len(receipts) == 0 {
			fmt.Println("Chain is empty.")
			return nil
		}

		report := receipt.VerifyChain(receipts)

		fmt.Printf("Chain: %s\n", sessionID)
		fmt.Printf("Receipts: %d\n", report.ReceiptCount)
		if !report.EarliestTimestamp.IsZero() {
			fmt.Printf("Span: %s → %s\n",
				report.EarliestTimestamp.Local().Format(time.RFC822),
				report.LatestTimestamp.Local().Format(time.RFC822),
			)
		}
		if report.CloudCoSignCount > 0 {
			fmt.Printf("Cloud co-signatures: %d\n", report.CloudCoSignCount)
		}
		fmt.Println()

		for _, v := range report.Receipts {
			mark := colorGreen + "✓" + colorReset
			if !v.Valid {
				mark = colorRed + "✗" + colorReset
			}
			ctx := ""
			if v.ContextID != "" {
				ctx = "  ctx=" + v.ContextID[:8] + "…"
			}
			hash := receipts[v.Index].PayloadHash
			if len(hash) > 16 {
				hash = hash[:16]
			}
			fmt.Printf("  [%d] %s %-22s %s…%s\n", v.Index, mark, v.Phase, hash, ctx)
			for _, e := range v.Errors {
				fmt.Printf("       %s%s%s\n", colorRed, e, colorReset)
			}
		}

		if len(report.Warnings) > 0 {
			fmt.Println()
			for _, w := range report.Warnings {
				fmt.Printf("  %swarning:%s %s\n", colorYellow, colorReset, w)
			}
		}

		fmt.Println()
		if report.ChainIntact() {
			fmt.Printf("%schain intact%s\n", colorGreen, colorReset)
			return nil
		}

		fmt.Printf("%schain BROKEN at receipt %d%s\n", colorRed, report.BrokenAt, colorReset)
		for _, e := range report.Errors {
			fmt.Printf("  %s\n", e)
		}
		return fmt.Errorf("chain verification failed")
	},
}

var receiptShowCmd = &cobra.Command{
	Use:   "show <session-id>",
	Short: "Pretty-print receipts in a chain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionID := args[0]
		dir := filepath.Join(confireDir(), "receipts")

		var path string
		if strings.HasSuffix(sessionID, ".jsonl") {
			path = sessionID
		} else {
			path = filepath.Join(dir, sessionID+".jsonl")
		}

		receipts, err := receipt.ReadChainFromFile(path)
		if err != nil {
			return fmt.Errorf("cannot read chain %q: %w", sessionID, err)
		}
		if len(receipts) == 0 {
			fmt.Println("Chain is empty.")
			return nil
		}

		for i, r := range receipts {
			data, err := json.MarshalIndent(r, "", "  ")
			if err != nil {
				return err
			}
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("── receipt %d ──\n%s\n", i, data)
		}
		return nil
	},
}

func init() {
	receiptCmd.AddCommand(receiptListCmd)
	receiptCmd.AddCommand(receiptVerifyCmd)
	receiptCmd.AddCommand(receiptShowCmd)
	rootCmd.AddCommand(receiptCmd)
}
