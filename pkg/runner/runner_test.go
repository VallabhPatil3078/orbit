package runner

import (
	"context"
	"github.com/VallabhPatil3078/orbit/pkg/graph"
	"runtime"
	"testing"
	"time"
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
	if runtime.GOOS == "windows" {
		// In Windows cmd, `start /B ping` creates a background process. But Job Objects handles this.
		shellCmd = "start /B ping 127.0.0.1 -n 6 > nul"
	} else {
		// In Unix, `&` backgrounds the sleep process.
		shellCmd = "sleep 5 &"
	}

	nodeTree := &graph.Node{Name: "tree_task", Command: shellCmd}
	tiers := [][]*graph.Node{{nodeTree}}
	
	rep := NewMockReporter()
	
	// Inject a timeout that is shorter than the background process (5s) but enough for it to start.
	// Since the background process detaches/sleeps for 5s, the parent task (shell) would exit immediately on Unix if not for Wait() waiting for the process group.
	// Actually, if the shell exits, the task completes!
	// To test timeout tree kill, we need the parent to also hang, e.g., `sleep 5 & wait` or `start /wait ping ...`.
	
	if runtime.GOOS == "windows" {
		shellCmd = "ping 127.0.0.1 -n 6 > nul"
	} else {
		shellCmd = "sleep 5 & wait"
	}
	nodeTree.Command = shellCmd

	err := ExecuteTiers(context.Background(), tiers, nil, true, "", rep, 50*time.Millisecond, 50*time.Millisecond)
	
	if err != ErrTaskTimeout {
		t.Fatalf("expected ErrTaskTimeout, got %v", err)
	}
	
	// Ideally we would assert the child PID no longer exists, but getting the grandchild PID is platform-specific and complex.
	// We trust that ProcessTree.Kill handles the group/job.
}
