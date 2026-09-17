package targets

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/BrunoTeixeira1996/nutgmonitor/internal/logger"
)

func TestMain(m *testing.M) {
	if err := logger.Setup("logs"); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestValidateSSHKeys(t *testing.T) {
	t.Run("empty key path is always valid", func(t *testing.T) {
		tgt := Target{SSHKey: ""}
		if err := tgt.ValidateSSHKeys(); err != nil {
			t.Fatalf("expected no error for empty SSHKey, got %v", err)
		}
	})

	t.Run("existing key path is valid", func(t *testing.T) {
		f, err := os.CreateTemp(t.TempDir(), "key")
		if err != nil {
			t.Fatalf("failed to create temp key file: %v", err)
		}
		tgt := Target{SSHKey: f.Name()}
		if err := tgt.ValidateSSHKeys(); err != nil {
			t.Fatalf("expected no error for an existing SSHKey, got %v", err)
		}
	})

	t.Run("missing key path is invalid", func(t *testing.T) {
		tgt := Target{SSHKey: "/nonexistent/path/does-not-exist"}
		if err := tgt.ValidateSSHKeys(); err == nil {
			t.Fatal("expected an error for a missing SSHKey, got nil")
		}
	})
}

func TestShutdownTargets(t *testing.T) {
	var mu sync.Mutex
	var called []string

	makeFunc := func(name string, delay time.Duration, fail bool) func(sshkey, ip string) error {
		return func(sshkey, ip string) error {
			time.Sleep(delay)
			mu.Lock()
			called = append(called, name)
			mu.Unlock()
			if fail {
				return fmt.Errorf("simulated shutdown error for %s", name)
			}
			return nil
		}
	}

	sleepEach := 150 * time.Millisecond
	ts := []Target{
		{Name: "a", ShutdownFunc: makeFunc("a", sleepEach, false)},
		{Name: "b", ShutdownFunc: makeFunc("b", sleepEach, true)}, // errors on "shutdown" (e.g. connection reset), must still be treated as attempted
		{Name: "pinute", ShutdownFunc: makeFunc("pinute", 0, false)},
	}

	start := time.Now()
	ShutdownTargets(ts)
	elapsed := time.Since(start)

	mu.Lock()
	defer mu.Unlock()

	if len(called) != 2 {
		t.Fatalf("expected 2 targets to be shut down (pinute skipped), got %d: %v", len(called), called)
	}
	for _, name := range called {
		if name == "pinute" {
			t.Fatal("pinute must never be shut down by ShutdownTargets - it is always shut down last, separately")
		}
	}

	// If targets were shut down sequentially this would take at least
	// 2*sleepEach. Running them concurrently should take close to a single
	// sleepEach, regardless of one of them returning an error.
	if elapsed >= 2*sleepEach {
		t.Fatalf("ShutdownTargets appears to run sequentially: took %v for 2 targets sleeping %v each", elapsed, sleepEach)
	}
}
