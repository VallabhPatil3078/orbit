package runner

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

var (
	iconSuccess       = "✔"
	iconFail          = "✖"
	iconStart         = "🪐"
	iconFinishSuccess = "✨"
	iconFinishFail    = "❌"
)

func init() {
	if runtime.GOOS == "windows" && os.Getenv("WT_SESSION") == "" {
		iconSuccess = "[OK]"
		iconFail    = "[FAIL]"
		iconStart   = "->"
		iconFinishSuccess = "==="
		iconFinishFail    = "[X]"
	}
}

// Reporter handles all CLI output and event logging for the Orbit pipeline.
type Reporter interface {
	PipelineStarted()
	TierStarted(tierIndex int, taskNames []string)
	TaskStarted(task string)
	TaskSkipped(task string, reason string)
	TaskSucceeded(task string)
	TaskFailed(task string, output string)
	PipelineFinished(success bool)
}

// TextReporter is the default implementation that prints standard text logs.
type TextReporter struct{}

func (r *TextReporter) PipelineStarted() {
	fmt.Printf("%s Starting Orbit Pipeline...\n", iconStart)
}

func (r *TextReporter) TierStarted(tierIndex int, taskNames []string) {
	fmt.Printf("\n[ Orbit ] Running %s...\n", strings.Join(taskNames, ", "))
}

func (r *TextReporter) TaskStarted(task string) {
	// TextReporter doesn't print on start, only on completion
}

func (r *TextReporter) TaskSkipped(task string, reason string) {
	fmt.Printf("  - %s (skipped: %s)\n", task, reason)
}

func (r *TextReporter) TaskSucceeded(task string) {
	fmt.Printf("  %s %s (completed)\n", iconSuccess, task)
}

func (r *TextReporter) TaskFailed(task string, output string) {
	fmt.Printf("\n  %s %s (failed):\n%s\n", iconFail, task, output)
}

func (r *TextReporter) PipelineFinished(success bool) {
	if success {
		fmt.Printf("\n%s Orbit pipeline completed successfully!\n", iconFinishSuccess)
	} else {
		fmt.Printf("\n%s Orbit pipeline failed!\n", iconFinishFail)
	}
}
