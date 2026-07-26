package azure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ankurCES/floc-zure/pkg/models"
	"github.com/ankurCES/floc-zure/pkg/utils"
)

// CLIExecutorImpl shells out to the `az` binary.
type CLIExecutorImpl struct {
	azPath string
}

// NewCLIExecutorImpl creates an executor that uses `az` from PATH.
// If the AZFLOCI_AZ_PATH env var is set, that path is used instead,
// enabling drop-in replacement with the Azure simulator binary.
func NewCLIExecutorImpl() *CLIExecutorImpl {
	p := os.Getenv("AZFLOCI_AZ_PATH")
	if p == "" {
		p, _ = exec.LookPath("az")
		if p == "" {
			p = "az"
		}
	}
	return &CLIExecutorImpl{azPath: p}
}

func (e *CLIExecutorImpl) Run(ctx context.Context, args ...string) (*CLIResult, error) {
	cmd := exec.CommandContext(ctx, e.azPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := &CLIResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	if err != nil {
		return result, &utils.CLIError{
			Command:  fmt.Sprintf("az %s", strings.Join(args, " ")),
			ExitCode: result.ExitCode,
			Stderr:   result.Stderr,
		}
	}
	return result, nil
}

func (e *CLIExecutorImpl) RunJSON(ctx context.Context, dest interface{}, args ...string) error {
	// Force JSON output
	args = append(args, "--output", "json")
	result, err := e.Run(ctx, args...)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(result.Stdout), dest)
}

func (e *CLIExecutorImpl) GetAccount(ctx context.Context) (*models.AzureAccount, error) {
	var acct models.AzureAccount
	if err := e.RunJSON(ctx, &acct, "account", "show"); err != nil {
		return nil, &utils.AuthError{Message: fmt.Sprintf("failed to get account: %v", err)}
	}
	return &acct, nil
}

func (e *CLIExecutorImpl) IsAuthenticated(ctx context.Context) (bool, error) {
	_, err := e.GetAccount(ctx)
	return err == nil, err
}

func (e *CLIExecutorImpl) SetSubscription(ctx context.Context, subscriptionID string) error {
	_, err := e.Run(ctx, "account", "set", "--subscription", subscriptionID)
	return err
}
