package cli

import (
	"context"
	"fmt"

	"github.com/ankurCES/floc-zure/internal/azure"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Azure authentication commands",
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current Azure account and auth status",
	RunE: func(cmd *cobra.Command, args []string) error {
		exec := azure.NewCLIExecutorImpl()
		ctx := context.Background()

		acct, err := exec.GetAccount(ctx)
		if err != nil {
			return fmt.Errorf("not authenticated: %w\nRun 'az login' to authenticate", err)
		}

		fmt.Printf("Authenticated ✓\n")
		fmt.Printf("  Subscription: %s (%s)\n", acct.Name, acct.ID)
		fmt.Printf("  Tenant:       %s\n", acct.TenantID)
		fmt.Printf("  User:         %s (%s)\n", acct.User.Name, acct.User.Type)
		return nil
	},
}

func init() {
	authCmd.AddCommand(authStatusCmd)
}
