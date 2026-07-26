package handlers

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ankurCES/floc-zure/simulator/internal/router"
	"github.com/ankurCES/floc-zure/simulator/internal/state"
)

func TestFACreateShowListDelete(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterFunctionAppHandlers(rtr, store)

	out := captureStdout(t, func() {
		if c := rtr.Dispatch([]string{"functionapp", "create", "--name", "fa1", "--resource-group", "rg1", "--runtime", "python"}); c != 0 {
			t.Fatalf("exit %d", c)
		}
	})
	var fa state.FunctionApp
	json.Unmarshal([]byte(out), &fa)
	if fa.Name != "fa1" || fa.Runtime != "python" {
		t.Errorf("got %+v", fa)
	}

	out = captureStdout(t, func() { rtr.Dispatch([]string{"functionapp", "show", "--name", "fa1"}) })
	if !strings.Contains(out, "fa1") {
		t.Error("show missing fa1")
	}

	out = captureStdout(t, func() { rtr.Dispatch([]string{"functionapp", "list"}) })
	var list []state.FunctionApp
	json.Unmarshal([]byte(out), &list)
	if len(list) != 1 {
		t.Errorf("list: %d", len(list))
	}

	if c := rtr.Dispatch([]string{"functionapp", "delete", "--name", "fa1"}); c != 0 {
		t.Fatalf("delete exit %d", c)
	}
}

func TestFuncCreateShowListDelete(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterFunctionAppHandlers(rtr, store)
	store.CreateFunctionApp("fa1", "rg1", "eastus", "node", "18", nil)

	out := captureStdout(t, func() {
		if c := rtr.Dispatch([]string{"functionapp", "function", "create", "--name", "fn1", "--function-app-name", "fa1", "--trigger-type", "httpTrigger"}); c != 0 {
			t.Fatalf("exit %d", c)
		}
	})
	if !strings.Contains(out, "fn1") {
		t.Error("missing fn1")
	}

	out = captureStdout(t, func() {
		rtr.Dispatch([]string{"functionapp", "function", "show", "--name", "fn1", "--function-app-name", "fa1"})
	})
	if !strings.Contains(out, "httpTrigger") {
		t.Error("show missing trigger")
	}

	out = captureStdout(t, func() {
		rtr.Dispatch([]string{"functionapp", "function", "list", "--function-app-name", "fa1"})
	})
	var fns []state.Function
	json.Unmarshal([]byte(out), &fns)
	if len(fns) != 1 {
		t.Errorf("list: %d", len(fns))
	}

	if c := rtr.Dispatch([]string{"functionapp", "function", "delete", "--name", "fn1", "--function-app-name", "fa1"}); c != 0 {
		t.Fatalf("delete exit %d", c)
	}
}

func TestFuncInvokeAndHistory(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterFunctionAppHandlers(rtr, store)
	store.CreateFunctionApp("fa1", "rg1", "eastus", "node", "18", nil)
	store.CreateFunction("fa1", "fn1", "httpTrigger", "node", nil)

	out := captureStdout(t, func() {
		if c := rtr.Dispatch([]string{"functionapp", "function", "invoke", "--name", "fn1", "--function-app-name", "fa1", "--input", `{"x":1}`}); c != 0 {
			t.Fatalf("invoke exit %d", c)
		}
	})
	var inv state.FunctionInvocation
	json.Unmarshal([]byte(out), &inv)
	if inv.Status != "Succeeded" {
		t.Errorf("status: %s", inv.Status)
	}

	out = captureStdout(t, func() {
		rtr.Dispatch([]string{"functionapp", "function", "invocations", "--function-app-name", "fa1"})
	})
	var invs []state.FunctionInvocation
	json.Unmarshal([]byte(out), &invs)
	if len(invs) != 1 {
		t.Errorf("invocations: %d", len(invs))
	}
}
