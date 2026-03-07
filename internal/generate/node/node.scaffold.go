package node

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ScaffoldTsoa writes a tsoa.json with sensible defaults to projectDir and
// returns the path of the file written.
func ScaffoldTsoa(projectDir string) (string, error) {
	outPath := filepath.Join(projectDir, "tsoa.json")

	entryFile := detectEntryFile(projectDir)

	cfg := map[string]any{
		"entryFile":                      entryFile,
		"noImplicitAdditionalProperties": "throw",
		"controllerPathGlobs":            []string{"src/**/*.controller.ts"},
		"spec": map[string]any{
			"outputDirectory": "build",
			"specVersion":     3,
		},
		"routes": map[string]any{
			"routesDir": "build",
		},
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal tsoa config: %w", err)
	}

	if err := os.WriteFile(outPath, append(data, '\n'), 0o644); err != nil {
		return "", fmt.Errorf("write tsoa.json: %w", err)
	}
	return outPath, nil
}

// InstallTsoa runs the appropriate package manager install command to add tsoa
// as a dev dependency.
func InstallTsoa(projectDir string) error {
	pm := detectPackageManager(projectDir)
	var args []string
	switch pm {
	case "pnpm":
		args = []string{"add", "--save-dev", "tsoa"}
	case "yarn":
		args = []string{"add", "--dev", "tsoa"}
	default:
		args = []string{"install", "--save-dev", "tsoa"}
		pm = "npm"
	}
	cmd := exec.Command(pm, args...)
	cmd.Dir = projectDir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s install tsoa: %w", pm, err)
	}
	return nil
}

// HasTsoaControllers reports whether the project at projectDir uses tsoa
// controller decorators (@Route, @Get, etc.). Returns false for plain Express
// projects that don't use tsoa's decorator-based approach.
func HasTsoaControllers(projectDir string) bool {
	srcDir := filepath.Join(projectDir, "src")
	if _, err := os.Stat(srcDir); err != nil {
		srcDir = projectDir
	}
	found := false
	_ = filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".ts" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if strings.Contains(string(data), "@Route(") {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func detectEntryFile(projectDir string) string {
	candidates := []string{
		"src/app.ts", "src/main.ts", "src/server.ts", "src/index.ts",
		"app.ts", "main.ts", "server.ts", "index.ts",
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(projectDir, c)); err == nil {
			return c
		}
	}
	return "src/app.ts"
}

func detectPackageManager(projectDir string) string {
	if _, err := os.Stat(filepath.Join(projectDir, "pnpm-lock.yaml")); err == nil {
		return "pnpm"
	}
	if _, err := os.Stat(filepath.Join(projectDir, "yarn.lock")); err == nil {
		return "yarn"
	}
	return "npm"
}
