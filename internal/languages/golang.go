package languages

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GenerateGo finds main.go in projectDir and runs swag init to produce an
// OpenAPI schema in outputDir.
func GenerateGo(projectDir, outputDir string) error {
	mainFile, err := findMainGo(projectDir)
	if err != nil {
		return err
	}

	rel, err := filepath.Rel(projectDir, mainFile)
	if err != nil {
		return fmt.Errorf("resolve main.go path: %w", err)
	}

	cmd := exec.Command("swag", "init", "--generalInfo", rel, "--output", outputDir, "--outputTypes", "yaml")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w\n\nHint: install swag with: go install github.com/swaggo/swag/cmd/swag@latest", err)
	}
	return nil
}

// findMainGo walks the project directory (up to 4 levels deep) to locate main.go,
// preferring cmd/*/ locations over others.
func findMainGo(dir string) (string, error) {
	var found []string
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == "node_modules" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			rel, _ := filepath.Rel(dir, path)
			if strings.Count(rel, string(filepath.Separator)) >= 4 {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "main.go" {
			found = append(found, path)
		}
		return nil
	})

	if len(found) == 0 {
		return "", fmt.Errorf(
			"cannot find main.go in project\n" +
				"Use --cmd to provide a custom generation command, e.g.:\n" +
				`  --cmd "swag init --generalInfo ./path/to/main.go" --output docs/swagger.yaml`,
		)
	}

	// prefer cmd/* locations
	for _, f := range found {
		if strings.Contains(f, string(filepath.Separator)+"cmd"+string(filepath.Separator)) {
			return f, nil
		}
	}
	return found[0], nil
}
