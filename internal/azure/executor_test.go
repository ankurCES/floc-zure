package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ankurCES/floc-zure/pkg/models"
	"github.com/ankurCES/floc-zure/pkg/utils"
)

// MockCLIExecutor implements CLIExecutor for testing without shelling out.
type MockCLIExecutor struct {
	RunFunc             func(ctx context.Context, args ...string) (*CLIResult, error)
	RunJSONFunc         func(ctx context.Context, dest interface{}, args ...string) error
	GetAccountFunc      func(ctx context.Context) (*models.AzureAccount, error)
	IsAuthenticatedFunc func(ctx context.Context) (bool, error)
	SetSubscriptionFunc func(ctx context.Context, id string) error
	// CallLog records every Run invocation for assertion.
	CallLog [][]string
}

func (m *MockCLIExecutor) Run(ctx context.Context, args ...string) (*CLIResult, error) {
	m.CallLog = append(m.CallLog, args)
	if m.RunFunc != nil {
		return m.RunFunc(ctx, args...)
	}
	return &CLIResult{Stdout: "{}", ExitCode: 0}, nil
}

func (m *MockCLIExecutor) RunJSON(ctx context.Context, dest interface{}, args ...string) error {
	m.CallLog = append(m.CallLog, args)
	if m.RunJSONFunc != nil {
		return m.RunJSONFunc(ctx, dest, args...)
	}
	return json.Unmarshal([]byte("{}"), dest)
}

func (m *MockCLIExecutor) GetAccount(ctx context.Context) (*models.AzureAccount, error) {
	if m.GetAccountFunc != nil {
		return m.GetAccountFunc(ctx)
	}
	return &models.AzureAccount{
		ID:       "sub-123",
		Name:     "Test Sub",
		TenantID: "tenant-456",
		User:     models.AzureUser{Name: "user@test.com", Type: "user"},
	}, nil
}

func (m *MockCLIExecutor) IsAuthenticated(ctx context.Context) (bool, error) {
	if m.IsAuthenticatedFunc != nil {
		return m.IsAuthenticatedFunc(ctx)
	}
	_, err := m.GetAccount(ctx)
	return err == nil, err
}

func (m *MockCLIExecutor) SetSubscription(ctx context.Context, id string) error {
	if m.SetSubscriptionFunc != nil {
		return m.SetSubscriptionFunc(ctx, id)
	}
	return nil
}

// --- Tests ---

func TestMockGetAccount(t *testing.T) {
	mock := &MockCLIExecutor{}
	acct, err := mock.GetAccount(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acct.ID != "sub-123" {
		t.Errorf("expected ID sub-123, got %s", acct.ID)
	}
	if acct.TenantID != "tenant-456" {
		t.Errorf("expected TenantID tenant-456, got %s", acct.TenantID)
	}
	if acct.User.Name != "user@test.com" {
		t.Errorf("expected user user@test.com, got %s", acct.User.Name)
	}
}

func TestMockIsAuthenticated_Success(t *testing.T) {
	mock := &MockCLIExecutor{}
	ok, err := mock.IsAuthenticated(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected authenticated=true")
	}
}

func TestMockIsAuthenticated_Failure(t *testing.T) {
	mock := &MockCLIExecutor{
		GetAccountFunc: func(ctx context.Context) (*models.AzureAccount, error) {
			return nil, &utils.AuthError{Message: "not logged in"}
		},
	}
	ok, err := mock.IsAuthenticated(context.Background())
	if ok {
		t.Error("expected authenticated=false")
	}
	if err == nil {
		t.Error("expected error")
	}
}

func TestMockSetSubscription(t *testing.T) {
	var captured string
	mock := &MockCLIExecutor{
		SetSubscriptionFunc: func(ctx context.Context, id string) error {
			captured = id
			return nil
		},
	}
	err := mock.SetSubscription(context.Background(), "new-sub-789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured != "new-sub-789" {
		t.Errorf("expected new-sub-789, got %s", captured)
	}
}

func TestMockRun_CallLog(t *testing.T) {
	mock := &MockCLIExecutor{}
	_, _ = mock.Run(context.Background(), "account", "show")
	_, _ = mock.Run(context.Background(), "group", "list")
	if len(mock.CallLog) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(mock.CallLog))
	}
	if mock.CallLog[0][0] != "account" {
		t.Errorf("expected first arg 'account', got %s", mock.CallLog[0][0])
	}
}

func TestMockRunJSON(t *testing.T) {
	mock := &MockCLIExecutor{
		RunJSONFunc: func(ctx context.Context, dest interface{}, args ...string) error {
			data := `{"id":"rg-1","name":"my-rg","location":"eastus"}`
			return json.Unmarshal([]byte(data), dest)
		},
	}
	var rg models.ResourceGroup
	err := mock.RunJSON(context.Background(), &rg, "group", "show", "--name", "my-rg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rg.Name != "my-rg" {
		t.Errorf("expected name my-rg, got %s", rg.Name)
	}
	if rg.Location != "eastus" {
		t.Errorf("expected location eastus, got %s", rg.Location)
	}
}

func TestMockRun_ErrorPropagation(t *testing.T) {
	mock := &MockCLIExecutor{
		RunFunc: func(ctx context.Context, args ...string) (*CLIResult, error) {
			return &CLIResult{ExitCode: 1, Stderr: "fail"}, fmt.Errorf("command failed")
		},
	}
	_, err := mock.Run(context.Background(), "bad", "cmd")
	if err == nil {
		t.Error("expected error from Run")
	}
}

func TestParseJSON(t *testing.T) {
	result := &CLIResult{Stdout: `{"name":"test-rg","location":"westus2"}`}
	rg, err := ParseJSON[models.ResourceGroup](result)
	if err != nil {
		t.Fatalf("ParseJSON failed: %v", err)
	}
	if rg.Name != "test-rg" {
		t.Errorf("expected name test-rg, got %s", rg.Name)
	}
	if rg.Location != "westus2" {
		t.Errorf("expected location westus2, got %s", rg.Location)
	}
}

func TestParseJSON_Invalid(t *testing.T) {
	result := &CLIResult{Stdout: `not json`}
	_, err := ParseJSON[models.ResourceGroup](result)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestCLIResult_Fields(t *testing.T) {
	r := &CLIResult{Stdout: "out", Stderr: "err", ExitCode: 42}
	if r.Stdout != "out" || r.Stderr != "err" || r.ExitCode != 42 {
		t.Error("CLIResult fields mismatch")
	}
}
