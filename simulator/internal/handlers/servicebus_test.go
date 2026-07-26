package handlers

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ankurCES/floc-zure/simulator/internal/router"
	"github.com/ankurCES/floc-zure/simulator/internal/state"
)

func TestSBNamespaceCreateShowListDelete(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterServiceBusHandlers(rtr, store)

	out := captureStdout(t, func() {
		if c := rtr.Dispatch([]string{"servicebus", "namespace", "create", "--name", "ns1", "--resource-group", "rg1"}); c != 0 {
			t.Fatalf("create exit %d", c)
		}
	})
	var ns state.ServiceBusNamespace
	json.Unmarshal([]byte(out), &ns)
	if ns.Name != "ns1" {
		t.Errorf("name: %s", ns.Name)
	}

	out = captureStdout(t, func() {
		rtr.Dispatch([]string{"servicebus", "namespace", "show", "--name", "ns1"})
	})
	if !strings.Contains(out, "ns1") {
		t.Error("show missing ns1")
	}

	out = captureStdout(t, func() {
		rtr.Dispatch([]string{"servicebus", "namespace", "list"})
	})
	var nsList []state.ServiceBusNamespace
	json.Unmarshal([]byte(out), &nsList)
	if len(nsList) != 1 {
		t.Errorf("list: %d", len(nsList))
	}

	if c := rtr.Dispatch([]string{"servicebus", "namespace", "delete", "--name", "ns1"}); c != 0 {
		t.Fatalf("delete exit %d", c)
	}
}

func TestSBQueueCRUD(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterServiceBusHandlers(rtr, store)
	store.CreateServiceBusNamespace("ns1", "rg1", "eastus", "Standard", nil)

	out := captureStdout(t, func() {
		if c := rtr.Dispatch([]string{"servicebus", "queue", "create", "--name", "q1", "--namespace-name", "ns1"}); c != 0 {
			t.Fatalf("exit %d", c)
		}
	})
	if !strings.Contains(out, "q1") {
		t.Error("missing q1")
	}

	out = captureStdout(t, func() {
		rtr.Dispatch([]string{"servicebus", "queue", "show", "--name", "q1", "--namespace-name", "ns1"})
	})
	if !strings.Contains(out, "q1") {
		t.Error("show missing q1")
	}

	if c := rtr.Dispatch([]string{"servicebus", "queue", "delete", "--name", "q1", "--namespace-name", "ns1"}); c != 0 {
		t.Fatalf("delete exit %d", c)
	}
}

func TestSBTopicAndSubCRUD(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterServiceBusHandlers(rtr, store)
	store.CreateServiceBusNamespace("ns1", "rg1", "eastus", "Standard", nil)

	captureStdout(t, func() {
		rtr.Dispatch([]string{"servicebus", "topic", "create", "--name", "t1", "--namespace-name", "ns1"})
	})
	captureStdout(t, func() {
		if c := rtr.Dispatch([]string{"servicebus", "topic", "subscription", "create", "--name", "s1", "--namespace-name", "ns1", "--topic-name", "t1"}); c != 0 {
			t.Fatalf("sub create exit %d", c)
		}
	})

	out := captureStdout(t, func() {
		rtr.Dispatch([]string{"servicebus", "topic", "subscription", "list", "--namespace-name", "ns1", "--topic-name", "t1"})
	})
	if !strings.Contains(out, "s1") {
		t.Error("sub list missing s1")
	}

	if c := rtr.Dispatch([]string{"servicebus", "topic", "subscription", "delete", "--name", "s1", "--namespace-name", "ns1", "--topic-name", "t1"}); c != 0 {
		t.Fatalf("sub delete exit %d", c)
	}
}

func TestSBMessageSendReceivePeek(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterServiceBusHandlers(rtr, store)
	store.CreateServiceBusNamespace("ns1", "rg1", "eastus", "Standard", nil)
	store.CreateServiceBusQueue("ns1", "q1", 0)

	captureStdout(t, func() {
		if c := rtr.Dispatch([]string{"servicebus", "queue", "message", "send", "--namespace-name", "ns1", "--queue-name", "q1", "--body", "hello"}); c != 0 {
			t.Fatalf("send exit %d", c)
		}
	})

	out := captureStdout(t, func() {
		rtr.Dispatch([]string{"servicebus", "queue", "message", "peek", "--namespace-name", "ns1", "--queue-name", "q1"})
	})
	if !strings.Contains(out, "hello") {
		t.Error("peek missing body")
	}

	out = captureStdout(t, func() {
		rtr.Dispatch([]string{"servicebus", "queue", "message", "receive", "--namespace-name", "ns1", "--queue-name", "q1"})
	})
	if !strings.Contains(out, "hello") {
		t.Error("receive missing body")
	}

	// empty queue returns null
	out = captureStdout(t, func() {
		rtr.Dispatch([]string{"servicebus", "queue", "message", "receive", "--namespace-name", "ns1", "--queue-name", "q1"})
	})
	if !strings.Contains(out, "null") {
		t.Error("expected null for empty queue")
	}
}
