package handlers

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ankurCES/floc-zure/simulator/internal/router"
	"github.com/ankurCES/floc-zure/simulator/internal/state"
)

// RegisterServiceBusHandlers wires servicebus commands into the router.
func RegisterServiceBusHandlers(r *router.Router, store *state.Store) {
	// az servicebus namespace ...
	r.Register("servicebus namespace create", sbNSCreate(store))
	r.Register("servicebus namespace show", sbNSShow(store))
	r.Register("servicebus namespace list", sbNSList(store))
	r.Register("servicebus namespace delete", sbNSDelete(store))

	// az servicebus queue ...
	r.Register("servicebus queue create", sbQueueCreate(store))
	r.Register("servicebus queue show", sbQueueShow(store))
	r.Register("servicebus queue list", sbQueueList(store))
	r.Register("servicebus queue delete", sbQueueDelete(store))

	// az servicebus topic ...
	r.Register("servicebus topic create", sbTopicCreate(store))
	r.Register("servicebus topic show", sbTopicShow(store))
	r.Register("servicebus topic list", sbTopicList(store))
	r.Register("servicebus topic delete", sbTopicDelete(store))

	// az servicebus topic subscription ...
	r.Register("servicebus topic subscription create", sbSubCreate(store))
	r.Register("servicebus topic subscription show", sbSubShow(store))
	r.Register("servicebus topic subscription list", sbSubList(store))
	r.Register("servicebus topic subscription delete", sbSubDelete(store))

	// message operations
	r.Register("servicebus queue message send", sbMsgSend(store))
	r.Register("servicebus queue message receive", sbMsgReceive(store))
	r.Register("servicebus queue message peek", sbMsgPeek(store))
}

// ---------------------------------------------------------------------------
// Namespace handlers
// ---------------------------------------------------------------------------

func sbNSCreate(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		rg, _ := router.ParseFlag(args, "--resource-group")
		if rg == "" {
			rg, _ = router.ParseFlag(args, "-g")
		}
		if rg == "" {
			fmt.Fprintln(os.Stderr, "ERROR: --resource-group is required.")
			return 2
		}
		location, _ := router.ParseFlag(args, "--location")
		if location == "" {
			location, _ = router.ParseFlag(args, "-l")
		}
		if location == "" {
			location = "eastus"
		}
		sku, _ := router.ParseFlag(args, "--sku")
		if sku == "" {
			sku = "Standard"
		}
		tags := router.ParseTags(args)
		if len(tags) == 0 {
			tags = nil
		}
		ns, err := store.CreateServiceBusNamespace(name, rg, location, sku, tags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(ns)
	}
}

func sbNSShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		ns := store.GetServiceBusNamespace(name)
		if ns == nil {
			fmt.Fprintf(os.Stderr, "ERROR: Namespace '%s' not found.\n", name)
			return 1
		}
		return writeJSON(ns)
	}
}

func sbNSList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		rg, _ := router.ParseFlag(args, "--resource-group")
		if rg == "" {
			rg, _ = router.ParseFlag(args, "-g")
		}
		return writeJSON(store.ListServiceBusNamespaces(rg))
	}
}

func sbNSDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		if err := store.DeleteServiceBusNamespace(name); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}

// ---------------------------------------------------------------------------
// Queue handlers
// ---------------------------------------------------------------------------

func sbQueueCreate(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		nsName, ok := router.ParseFlag(args, "--namespace-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --namespace-name is required.")
			return 2
		}
		maxSizeStr, _ := router.ParseFlag(args, "--max-size")
		maxSize := 1024
		if maxSizeStr != "" {
			if v, err := strconv.Atoi(maxSizeStr); err == nil {
				maxSize = v
			}
		}
		q, err := store.CreateServiceBusQueue(nsName, name, maxSize)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(q)
	}
}

func sbQueueShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		nsName, ok := router.ParseFlag(args, "--namespace-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --namespace-name is required.")
			return 2
		}
		q := store.GetServiceBusQueue(nsName, name)
		if q == nil {
			fmt.Fprintf(os.Stderr, "ERROR: Queue '%s' not found in namespace '%s'.\n", name, nsName)
			return 1
		}
		return writeJSON(q)
	}
}

func sbQueueList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		nsName, ok := router.ParseFlag(args, "--namespace-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --namespace-name is required.")
			return 2
		}
		return writeJSON(store.ListServiceBusQueues(nsName))
	}
}

func sbQueueDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		nsName, ok := router.ParseFlag(args, "--namespace-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --namespace-name is required.")
			return 2
		}
		if err := store.DeleteServiceBusQueue(nsName, name); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}

// ---------------------------------------------------------------------------
// Topic handlers
// ---------------------------------------------------------------------------

func sbTopicCreate(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		nsName, ok := router.ParseFlag(args, "--namespace-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --namespace-name is required.")
			return 2
		}
		maxSizeStr, _ := router.ParseFlag(args, "--max-size")
		maxSize := 1024
		if maxSizeStr != "" {
			if v, err := strconv.Atoi(maxSizeStr); err == nil {
				maxSize = v
			}
		}
		t, err := store.CreateServiceBusTopic(nsName, name, maxSize)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(t)
	}
}

func sbTopicShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		nsName, ok := router.ParseFlag(args, "--namespace-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --namespace-name is required.")
			return 2
		}
		t := store.GetServiceBusTopic(nsName, name)
		if t == nil {
			fmt.Fprintf(os.Stderr, "ERROR: Topic '%s' not found in namespace '%s'.\n", name, nsName)
			return 1
		}
		return writeJSON(t)
	}
}

func sbTopicList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		nsName, ok := router.ParseFlag(args, "--namespace-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --namespace-name is required.")
			return 2
		}
		return writeJSON(store.ListServiceBusTopics(nsName))
	}
}

func sbTopicDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		nsName, ok := router.ParseFlag(args, "--namespace-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --namespace-name is required.")
			return 2
		}
		if err := store.DeleteServiceBusTopic(nsName, name); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}

// ---------------------------------------------------------------------------
// Subscription handlers
// ---------------------------------------------------------------------------

func sbSubCreate(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		nsName, ok := router.ParseFlag(args, "--namespace-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --namespace-name is required.")
			return 2
		}
		topicName, ok := router.ParseFlag(args, "--topic-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --topic-name is required.")
			return 2
		}
		maxDeliveryStr, _ := router.ParseFlag(args, "--max-delivery-count")
		maxDelivery := 10
		if maxDeliveryStr != "" {
			if v, err := strconv.Atoi(maxDeliveryStr); err == nil {
				maxDelivery = v
			}
		}
		sub, err := store.CreateServiceBusSub(nsName, topicName, name, maxDelivery)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(sub)
	}
}

func sbSubShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		nsName, ok := router.ParseFlag(args, "--namespace-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --namespace-name is required.")
			return 2
		}
		topicName, ok := router.ParseFlag(args, "--topic-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --topic-name is required.")
			return 2
		}
		sub := store.GetServiceBusSub(nsName, topicName, name)
		if sub == nil {
			fmt.Fprintf(os.Stderr, "ERROR: Subscription '%s' not found on topic '%s'.\n", name, topicName)
			return 1
		}
		return writeJSON(sub)
	}
}

func sbSubList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		nsName, ok := router.ParseFlag(args, "--namespace-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --namespace-name is required.")
			return 2
		}
		topicName, ok := router.ParseFlag(args, "--topic-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --topic-name is required.")
			return 2
		}
		return writeJSON(store.ListServiceBusSubs(nsName, topicName))
	}
}

func sbSubDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		nsName, ok := router.ParseFlag(args, "--namespace-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --namespace-name is required.")
			return 2
		}
		topicName, ok := router.ParseFlag(args, "--topic-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --topic-name is required.")
			return 2
		}
		if err := store.DeleteServiceBusSub(nsName, topicName, name); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}

// ---------------------------------------------------------------------------
// Message handlers
// ---------------------------------------------------------------------------

func sbMsgSend(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		nsName, ok := router.ParseFlag(args, "--namespace-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --namespace-name is required.")
			return 2
		}
		queueName, ok := router.ParseFlag(args, "--queue-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --queue-name is required.")
			return 2
		}
		body, ok := router.ParseFlag(args, "--body")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --body is required.")
			return 2
		}
		contentType, _ := router.ParseFlag(args, "--content-type")
		label, _ := router.ParseFlag(args, "--label")
		msg, err := store.SendMessage(nsName, queueName, body, contentType, label, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(msg)
	}
}

func sbMsgReceive(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		nsName, ok := router.ParseFlag(args, "--namespace-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --namespace-name is required.")
			return 2
		}
		queueName, ok := router.ParseFlag(args, "--queue-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --queue-name is required.")
			return 2
		}
		msg, err := store.ReceiveMessage(nsName, queueName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		if msg == nil {
			fmt.Println("null")
			return 0
		}
		return writeJSON(msg)
	}
}

func sbMsgPeek(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		nsName, ok := router.ParseFlag(args, "--namespace-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --namespace-name is required.")
			return 2
		}
		queueName, ok := router.ParseFlag(args, "--queue-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --queue-name is required.")
			return 2
		}
		msg, err := store.PeekMessage(nsName, queueName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		if msg == nil {
			fmt.Println("null")
			return 0
		}
		return writeJSON(msg)
	}
}
