package impact

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// skipDirs lists directory names that are never scanned.
var skipDirs = map[string]bool{
	"vendor":      true,
	"node_modules": true,
	".git":        true,
	".idea":       true,
	"dist":        true,
	"build":       true,
	"__pycache__": true,
}

// textExtensions is the set of file extensions treated as scannable text.
var textExtensions = map[string]bool{
	".go":      true,
	".ts":      true,
	".js":      true,
	".tsx":     true,
	".jsx":     true,
	".py":      true,
	".rb":      true,
	".java":    true,
	".kt":      true,
	".scala":   true,
	".cs":      true,
	".rs":      true,
	".cpp":     true,
	".c":       true,
	".h":       true,
	".swift":   true,
	".php":     true,
	".graphql": true,
	".gql":     true,
	".proto":   true,
	".yaml":    true,
	".yml":     true,
	".sh":      true,
	".bash":    true,
	".toml":    true,
}

// Scan walks dir recursively and returns all line-level hits for terms.
// changePath and changeType are attached to every Hit for grouping.
func Scan(dir string, terms []string, changePath, changeType string) ([]Hit, error) {
	if len(terms) == 0 {
		return nil, nil
	}
	var hits []Hit
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !isTextFile(path) {
			return nil
		}
		fileHits, ferr := scanFile(path, terms, changePath, changeType)
		if ferr != nil {
			return nil // skip unreadable files
		}
		hits = append(hits, fileHits...)
		return nil
	})
	return hits, err
}

func scanFile(path string, terms []string, changePath, changeType string) ([]Hit, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var hits []Hit
	sc := bufio.NewScanner(f)
	lineNum := 0
	for sc.Scan() {
		lineNum++
		line := sc.Text()
		for _, term := range terms {
			if strings.Contains(line, term) {
				hits = append(hits, Hit{
					File:       path,
					LineNum:    lineNum,
					Line:       strings.TrimSpace(line),
					ChangeType: changeType,
					ChangePath: changePath,
				})
				break // one hit per line per change
			}
		}
	}
	return hits, sc.Err()
}

func isTextFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return textExtensions[ext]
}
