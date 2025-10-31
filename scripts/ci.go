package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// TaskResult holds the result of a single task
type TaskResult struct {
	Name     string
	Success  bool
	Duration time.Duration
	Output   string
	Error    string
}

// filteringWriter filters out verbose output patterns
type filteringWriter struct {
	writer       io.Writer
	buffer       bytes.Buffer
	skipPatterns []string
	inSkipBlock  bool
}

func newFilteringWriter(w io.Writer) *filteringWriter {
	return &filteringWriter{
		writer: w,
		skipPatterns: []string{
			"Starting Anvil network...",
			"pkill -f",
			"anvil &",
			"Waiting for Anvil to be ready...",
			"_   _",
			"(_) | |",
			"__ _",
			"/ _`",
			"| (_|",
			"\\__,_",
			"https://github.com/foundry-rs/foundry",
			"Available Accounts",
			"==================",
			"Private Keys",
			"Wallet",
			"Mnemonic:",
			"Derivation path:",
			"Chain ID",
			"Base Fee",
			"Gas Limit",
			"Genesis Timestamp",
			"Genesis Number",
			"Listening on",
			"0x", // Account addresses and private keys
		},
	}
}

func (f *filteringWriter) Write(p []byte) (n int, err error) {
	scanner := bufio.NewScanner(bytes.NewReader(p))
	for scanner.Scan() {
		line := scanner.Text()
		shouldSkip := false

		// Check if line matches any skip pattern
		for _, pattern := range f.skipPatterns {
			if strings.Contains(line, pattern) {
				shouldSkip = true
				break
			}
		}

		// Also skip empty lines that are part of Anvil output
		trimmed := strings.TrimSpace(line)
		if trimmed == "" && f.inSkipBlock {
			shouldSkip = true
		}

		if shouldSkip {
			f.inSkipBlock = true
		} else {
			// If we see actual test output, stop skipping
			if strings.Contains(line, "Running tests...") ||
				strings.Contains(line, "?") ||
				strings.Contains(line, "ok") ||
				strings.Contains(line, "FAIL") ||
				strings.Contains(line, "PASS") {
				f.inSkipBlock = false
			}

			if !f.inSkipBlock {
				f.writer.Write([]byte(line + "\n"))
			}
		}
	}

	return len(p), scanner.Err()
}

// runTask executes a command and captures its output while showing filtered real-time progress
func runTask(name string, command string, args ...string) TaskResult {
	start := time.Now()
	result := TaskResult{
		Name: name,
	}

	cmd := exec.Command(command, args...)
	var stdout, stderr bytes.Buffer

	// Create filtering writers to suppress verbose output like Anvil banner
	stdoutFilter := newFilteringWriter(os.Stdout)
	stderrFilter := newFilteringWriter(os.Stderr)

	// Create multi-writers to capture output AND show filtered real-time output
	cmd.Stdout = io.MultiWriter(&stdout, stdoutFilter)
	cmd.Stderr = io.MultiWriter(&stderr, stderrFilter)

	err := cmd.Run()
	result.Duration = time.Since(start)
	result.Output = stdout.String()
	result.Error = stderr.String()
	result.Success = err == nil

	return result
}

// formatDuration formats duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", d.Seconds()*1000)
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

// printHeader prints a section header
func printHeader(text string) {
	line := strings.Repeat("═", len(text)+4)
	fmt.Printf("\n%s\n  %s\n%s\n", line, text, line)
}

// printTaskSummary prints a summary of a single task
func printTaskSummary(result TaskResult) {
	statusIcon := "✓"
	statusColor := "\033[32m" // green
	if !result.Success {
		statusIcon = "✗"
		statusColor = "\033[31m" // red
	}
	resetColor := "\033[0m"

	fmt.Printf("%s%s%s %s (%s)\n", statusColor, statusIcon, resetColor, result.Name, formatDuration(result.Duration))
}

// printTaskDetails prints detailed output for a task
func printTaskDetails(result TaskResult) {
	printHeader(fmt.Sprintf("%s Details", result.Name))

	if result.Success {
		fmt.Println("\n✓ Passed")
		if result.Output != "" {
			fmt.Println("\nOutput:")
			fmt.Println(result.Output)
		}
	} else {
		fmt.Println("\n✗ Failed")
		if result.Error != "" {
			fmt.Println("\nError output:")
			fmt.Println(result.Error)
		}
		if result.Output != "" {
			fmt.Println("\nStandard output:")
			fmt.Println(result.Output)
		}
	}
}

func main() {
	printHeader("Running CI Tasks")
	fmt.Println("Executing: make test, make lint, make build sequentially...")
	fmt.Println()

	tasks := []struct {
		name string
		cmd  string
		args []string
	}{
		{"Test", "make", []string{"test"}},
		{"Lint", "make", []string{"lint"}},
		{"Build", "make", []string{"build"}},
	}

	results := make([]TaskResult, len(tasks))

	// Run tasks sequentially
	overallStart := time.Now()
	for i, task := range tasks {
		fmt.Printf("Running %s...\n", task.name)
		results[i] = runTask(task.name, task.cmd, task.args...)

		// Stop on first failure for faster feedback
		if !results[i].Success {
			fmt.Printf("\n❌ %s failed, stopping CI run\n\n", task.name)
			// Fill remaining results with empty entries
			for j := i + 1; j < len(tasks); j++ {
				results[j] = TaskResult{
					Name:    tasks[j].name,
					Success: false,
					Output:  "Skipped due to previous failure",
				}
			}
			break
		}
	}

	overallDuration := time.Since(overallStart)

	// Print summary
	printHeader("Summary")
	fmt.Println()

	successCount := 0
	for _, result := range results {
		printTaskSummary(result)
		if result.Success {
			successCount++
		}
	}

	fmt.Printf("\nTotal time: %s\n", formatDuration(overallDuration))
	fmt.Printf("Passed: %d/%d\n", successCount, len(results))

	// Print details for failed tasks
	hasFailures := false
	for _, result := range results {
		if !result.Success {
			hasFailures = true
			printTaskDetails(result)
		}
	}

	// Print success message or failure details
	if !hasFailures {
		printHeader("All Checks Passed!")
		fmt.Println()
		os.Exit(0)
	} else {
		printHeader("Some Checks Failed")
		fmt.Println("\nReview the details above for more information.")
		fmt.Println()
		os.Exit(1)
	}
}
