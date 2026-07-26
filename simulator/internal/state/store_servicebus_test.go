package state

import (
	"path/filepath"
	"testing"
)

func sbStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

func TestCreateServiceBusNamespace(t *testing.T) {
	s := sbStore(t)
	ns, err := s.CreateServiceBusNamespace("ns1", "rg1", "eastus", "Standard", nil)
	if err != nil {
		t.Fatal(err)
	}
	if ns.Name != "ns1" || ns.Location != "eastus" {
		t.Errorf("got %+v", ns)
	}
	// duplicate
	if _, err := s.CreateServiceBusNamespace("ns1", "rg1", "eastus", "Standard", nil); err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestServiceBusNamespaceCRUD(t *testing.T) {
	s := sbStore(t)
	s.CreateServiceBusNamespace("ns1", "rg1", "eastus", "Standard", nil)
	if got := s.GetServiceBusNamespace("ns1"); got == nil {
		t.Fatal("expected ns1")
	}
	if got := s.ListServiceBusNamespaces(""); len(got) != 1 {
		t.Fatalf("expected 1, got %d", len(got))
	}
	if err := s.DeleteServiceBusNamespace("ns1"); err != nil {
		t.Fatal(err)
	}
	if got := s.GetServiceBusNamespace("ns1"); got != nil {
		t.Fatal("expected nil after delete")
	}
}

func TestServiceBusQueueCRUD(t *testing.T) {
	s := sbStore(t)
	s.CreateServiceBusNamespace("ns1", "rg1", "eastus", "Standard", nil)
	q, err := s.CreateServiceBusQueue("ns1", "q1", 0)
	if err != nil {
		t.Fatal(err)
	}
	if q.MaxSizeInMegabytes != 1024 {
		t.Errorf("default maxSize: %d", q.MaxSizeInMegabytes)
	}
	if got := s.GetServiceBusQueue("ns1", "q1"); got == nil {
		t.Fatal("expected q1")
	}
	if got := s.ListServiceBusQueues("ns1"); len(got) != 1 {
		t.Fatalf("expected 1, got %d", len(got))
	}
	if err := s.DeleteServiceBusQueue("ns1", "q1"); err != nil {
		t.Fatal(err)
	}
}

func TestServiceBusQueueNotFoundNamespace(t *testing.T) {
	s := sbStore(t)
	if _, err := s.CreateServiceBusQueue("nope", "q1", 0); err == nil {
		t.Fatal("expected error for missing namespace")
	}
}

func TestServiceBusTopicAndSubscription(t *testing.T) {
	s := sbStore(t)
	s.CreateServiceBusNamespace("ns1", "rg1", "eastus", "Standard", nil)
	topic, err := s.CreateServiceBusTopic("ns1", "t1", 2048)
	if err != nil {
		t.Fatal(err)
	}
	if topic.MaxSizeInMegabytes != 2048 {
		t.Errorf("maxSize: %d", topic.MaxSizeInMegabytes)
	}
	sub, err := s.CreateServiceBusSub("ns1", "t1", "sub1", 5)
	if err != nil {
		t.Fatal(err)
	}
	if sub.MaxDeliveryCount != 5 {
		t.Errorf("maxDelivery: %d", sub.MaxDeliveryCount)
	}
	// topic subscription count should be 1
	if got := s.GetServiceBusTopic("ns1", "t1"); got.SubscriptionCount != 1 {
		t.Errorf("subCount: %d", got.SubscriptionCount)
	}
	if got := s.ListServiceBusSubs("ns1", "t1"); len(got) != 1 {
		t.Fatalf("expected 1 sub, got %d", len(got))
	}
	if err := s.DeleteServiceBusSub("ns1", "t1", "sub1"); err != nil {
		t.Fatal(err)
	}
	if got := s.GetServiceBusTopic("ns1", "t1"); got.SubscriptionCount != 0 {
		t.Errorf("subCount after delete: %d", got.SubscriptionCount)
	}
}

func TestServiceBusMessages(t *testing.T) {
	s := sbStore(t)
	s.CreateServiceBusNamespace("ns1", "rg1", "eastus", "Standard", nil)
	s.CreateServiceBusQueue("ns1", "q1", 0)

	// send
	msg, err := s.SendMessage("ns1", "q1", "hello", "text/plain", "lbl", nil)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Body != "hello" {
		t.Errorf("body: %s", msg.Body)
	}

	// peek (non-destructive)
	peeked, _ := s.PeekMessage("ns1", "q1")
	if peeked == nil || peeked.Body != "hello" {
		t.Fatal("peek failed")
	}

	// receive (destructive)
	received, _ := s.ReceiveMessage("ns1", "q1")
	if received == nil || received.Body != "hello" {
		t.Fatal("receive failed")
	}

	// empty queue
	empty, _ := s.ReceiveMessage("ns1", "q1")
	if empty != nil {
		t.Fatal("expected nil for empty queue")
	}
}

func TestNamespaceDeleteCascade(t *testing.T) {
	s := sbStore(t)
	s.CreateServiceBusNamespace("ns1", "rg1", "eastus", "Standard", nil)
	s.CreateServiceBusQueue("ns1", "q1", 0)
	s.CreateServiceBusTopic("ns1", "t1", 0)
	s.CreateServiceBusSub("ns1", "t1", "sub1", 0)
	s.SendMessage("ns1", "q1", "test", "", "", nil)

	if err := s.DeleteServiceBusNamespace("ns1"); err != nil {
		t.Fatal(err)
	}
	if s.GetServiceBusQueue("ns1", "q1") != nil {
		t.Error("queue should be deleted")
	}
	if s.GetServiceBusTopic("ns1", "t1") != nil {
		t.Error("topic should be deleted")
	}
}
