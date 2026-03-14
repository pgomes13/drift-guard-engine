package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pgomes13/api-drift-engine/internal/languages"
)

// --------------------------------------------------------------------------
// progress spinner
// --------------------------------------------------------------------------

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// runStep displays an animated spinner while fn executes, then prints ✓ or ✗.
func runStep(label string, fn func() error) error {
	stop := make(chan struct{})
	spinDone := make(chan struct{})
	var fnErr error

	go func() {
		defer close(spinDone)
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		i := 0
		fmt.Fprintf(os.Stderr, "  %s  %s", spinnerFrames[0], label)
		for {
			select {
			case <-stop:
				if fnErr != nil {
					fmt.Fprintf(os.Stderr, "\r  ✗  %s\n", label)
				} else {
					fmt.Fprintf(os.Stderr, "\r  ✓  %s\n", label)
				}
				return
			case <-ticker.C:
				i++
				fmt.Fprintf(os.Stderr, "\r  %s  %s", spinnerFrames[i%len(spinnerFrames)], label)
			}
		}
	}()

	fnErr = fn()
	close(stop)
	<-spinDone
	return fnErr
}

// --------------------------------------------------------------------------
// interactive prompts
// --------------------------------------------------------------------------

// stdinReader is shared across all promptYesNo calls so that bufio buffering
// doesn't consume input intended for a subsequent prompt.
var stdinReader = bufio.NewReader(os.Stdin)

// promptYesNo prints prompt and reads a Y/n response from stdin.
// An empty response (just Enter) defaults to Yes.
func promptYesNo(prompt string) bool {
	fmt.Fprintf(os.Stderr, "%s [Y/n]: ", prompt)
	line, _ := stdinReader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "" || line == "y" || line == "yes"
}

// --------------------------------------------------------------------------
// API type detection
// --------------------------------------------------------------------------

// detectAPITypes returns the API types present in the project.
// REST is always included; GraphQL and gRPC are added when detected.
func detectAPITypes(dir string) []string {
	types := []string{"REST"}
	if languages.DetectGraphQLInfo(dir) != nil {
		types = append(types, "GraphQL")
	}
	if hasProtoFiles(dir) {
		types = append(types, "gRPC")
	}
	return types
}

// hasProtoFiles reports whether any .proto files exist under dir.
func hasProtoFiles(dir string) bool {
	found := false
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".proto") {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}
