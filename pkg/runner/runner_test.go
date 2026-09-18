package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/VallabhPatil3078/orbit/pkg/graph"
)

func TestExecuteTiers_Success(t *testing.T) {
	node1 := &graph.Node{Name: "echo1", Command: "echo 1"}
	node2 := &graph.Node{Name: "echo2", Command: "echo 2"}
	
	tiers := [][]*graph.Node{
		{node1, node2},
	}

	rep := NewMockReporter()
	err := ExecuteTiers(context.Background(), tiers, nil, true, "", rep, 10*time.Minute, 5*time.Second)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestExecuteTiers_Failure(t *testing.T) {
	nodeFail := &graph.Node{Name: "fail_task", Command: "exit 1"}
	tiers := [][]*graph.Node{
		{nodeFail},
	}
	rep := NewMockReporter()
	err := ExecuteTiers(context.Background(), tiers, nil, true, "", rep, 10*time.Minute, 5*time.Second)
	if err == nil {
		t.Fatal("expected an error because the task fails, got nil")
	}
}

func TestExecuteTiers_RunsConcurrently(t *testing.T) {
	var sleepCmd string
	if runtime.GOOS == "windows" {
		// Use ping instead of powershell to avoid slow startup times in CI causing false positive test failures
		sleepCmd = "ping 127.0.0.1 -n 2 > nul"
	} else {
		sleepCmd = "sleep 1"
	}

	node1 := &graph.Node{Name: "sleep1", Command: sleepCmd}
	node2 := &graph.Node{Name: "sleep2", Command: sleepCmd}
	
	tiers := [][]*graph.Node{
		{node1, node2},
	}

	start := time.Now()
	rep := NewMockReporter()
	err := ExecuteTiers(context.Background(), tiers, nil, true, "", rep, 10*time.Minute, 5*time.Second)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error executing tiers: %v", err)
	}

	if duration >= 3000*time.Millisecond {
		t.Errorf("expected execution time < 3.0s, got %v (implies tasks ran sequentially)", duration)
	}
}

func TestExecuteTiers_Timeout(t *testing.T) {
	var sleepCmd string
	if runtime.GOOS == "windows" {
		// Ping for ~6 seconds
		sleepCmd = "ping 127.0.0.1 -n 6 > nul"
	} else {
		sleepCmd = "sleep 5"
	}

	nodeSlow := &graph.Node{Name: "slow_task", Command: sleepCmd}
	tiers := [][]*graph.Node{{nodeSlow}}
	
	rep := NewMockReporter()
	
	// Inject a very short timeout (50ms)
	err := ExecuteTiers(context.Background(), tiers, nil, true, "", rep, 50*time.Millisecond, 50*time.Millisecond)
	
	if err != ErrTaskTimeout {
		t.Fatalf("expected ErrTaskTimeout, got %v", err)
	}
}

func TestExecuteTiers_ProcessTreeCleanup(t *testing.T) {
	// This tests that when Orbit kills a task, it kills the entire process tree, not just the parent.
	var shellCmd string
	var pidFile string
	if runtime.GOOS == "windows" {
		shellCmd = "ping 127.0.0.1 -n 6 > nul"
	} else {
		// Write the background process PID to a file so we can check it later
		pidFile = "test_grandchild.pid"
		shellCmd = fmt.Sprintf("sleep 5 & echo $! > %s; wait", pidFile)
		defer os.Remove(pidFile)
	}

	nodeTree := &graph.Node{Name: "tree_task", Command: shellCmd}
	tiers := [][]*graph.Node{{nodeTree}}
	
	rep := NewMockReporter()
	
	err := ExecuteTiers(context.Background(), tiers, nil, true, "", rep, 50*time.Millisecond, 50*time.Millisecond)
	
	if err != ErrTaskTimeout {
		t.Fatalf("expected ErrTaskTimeout, got %v", err)
	}
	
	if runtime.GOOS != "windows" {
		// Read the PID from the file
		pidBytes, err := os.ReadFile(pidFile)
		if err == nil && len(pidBytes) > 0 {
			pidStr := strings.TrimSpace(string(pidBytes))
			// Check if process is still alive using kill -0
			checkCmd := exec.Command("sh", "-c", fmt.Sprintf("kill -0 %s", pidStr))
			if checkCmd.Run() == nil {
				t.Fatalf("Grandchild process %s is still alive! Process tree was not cleaned up.", pidStr)
			}
		} else {
			t.Logf("Warning: Could not read pidfile, skipping assertion. Err: %v", err)
		}
	}
}

func TestDetermineTaskStatus(t *testing.T) {
	// 1. Normal success
	status, ctxErr := DetermineTaskStatus(nil, nil)
	if status != StatusSuccess || ctxErr != nil {
		t.Errorf("expected Success/nil, got %v/%v", status, ctxErr)
	}

	// 2. Race condition: waitErr is nil, but context timed out
	status, ctxErr = DetermineTaskStatus(nil, context.DeadlineExceeded)
	if status != StatusSuccess || ctxErr != nil {
		t.Errorf("expected Success/nil (race handled), got %v/%v", status, ctxErr)
	}

	// 3. Normal failure
	someErr := fmt.Errorf("exit status 1")
	status, ctxErr = DetermineTaskStatus(someErr, nil)
	if status != StatusFailed || ctxErr != nil {
		t.Errorf("expected Failed/nil, got %v/%v", status, ctxErr)
	}

	// 4. Timeout failure
	status, ctxErr = DetermineTaskStatus(someErr, context.DeadlineExceeded)
	if status != StatusFailed || ctxErr != context.DeadlineExceeded {
		t.Errorf("expected Failed/DeadlineExceeded, got %v/%v", status, ctxErr)
	}
}
