package main

import (
	"fmt"
	"maunium.net/go/mautrix/crypto"
	"strings"
)

// Simple crypto.Logger implementation that just prints to stdout.
type fakeLogger struct{}

var _ crypto.Logger = &fakeLogger{}

func (f fakeLogger) Error(message string, args ...interface{}) {
	fmt.Printf("[ERROR] "+message+"\n", args...)
}

func (f fakeLogger) Warn(message string, args ...interface{}) {
	fmt.Printf("[WARN] "+message+"\n", args...)
}

func (f fakeLogger) Debug(message string, args ...interface{}) {
	fmt.Printf("[DEBUG] "+message+"\n", args...)
}

func (f fakeLogger) Trace(message string, args ...interface{}) {
	if strings.HasPrefix(message, "Got membership state event") {
		return
	}
	fmt.Printf("[TRACE] "+message+"\n", args...)
}
