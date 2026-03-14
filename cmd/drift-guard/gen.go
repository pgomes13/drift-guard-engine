package main

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/pgomes13/api-drift-engine/internal/languages"
)

// runGenerate auto-detects the project type and generates an OpenAPI schema
// for the project at projectDir, writing output files into outputDir.
func runGenerate(projectDir, outputDir string) error {
	gen, err := languages.DetectGenerator(projectDir)
	if err != nil {
		return err
	}
	return gen(projectDir, outputDir)
}

// runGraphQLGenerate auto-detects the project type and generates a GraphQL schema
// for the project at projectDir, writing output files into outputDir.
func runGraphQLGenerate(projectDir, outputDir string) error {
	gen, err := languages.DetectGraphQLGenerator(projectDir)
	if err != nil {
		return err
	}
	return gen(projectDir, outputDir)
}

// runGRPCGenerate auto-detects the project type and collects the gRPC proto schema
// for the project at projectDir, writing output files into outputDir.
func runGRPCGenerate(projectDir, outputDir string) error {
	gen, err := languages.DetectGRPCGenerator(projectDir)
	if err != nil {
		return err
	}
	return gen(projectDir, outputDir)
}

// addWorktree creates a detached git worktree at baseRef inside tmpDir and
// returns the worktree path along with a cleanup function.
func addWorktree(tmpDir, baseRef string) (string, func(), error) {
	dir := filepath.Join(tmpDir, "worktree")
	out, err := exec.Command("git", "worktree", "add", "--detach", dir, baseRef).CombinedOutput()
	if err != nil {
		return "", nil, fmt.Errorf("git worktree add %s: %w\n%s", baseRef, err, out)
	}
	cleanup := func() {
		exec.Command("git", "worktree", "remove", "--force", dir).Run()
	}
	return dir, cleanup, nil
}

// resolveBaseRef returns the first of the candidates that exists as a valid git ref.
func resolveBaseRef(baseRef string) string {
	if refExists(baseRef) {
		return baseRef
	}
	for _, candidate := range []string{"origin/master", "HEAD~1"} {
		if refExists(candidate) {
			return candidate
		}
	}
	return baseRef
}

func refExists(ref string) bool {
	return exec.Command("git", "rev-parse", "--verify", ref).Run() == nil
}
