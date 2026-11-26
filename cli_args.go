package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// extractMediaArgPaths normalizes CLI or second-instance file arguments into
// unique, existing non-directory file paths, ignoring flags and the executable itself.
func extractMediaArgPaths(args []string, workingDir string, selfExePath string) []string {
	var normalizedSelf string
	if strings.TrimSpace(selfExePath) != "" {
		if absSelf, err := filepath.Abs(filepath.Clean(selfExePath)); err == nil {
			normalizedSelf = absSelf
		} else {
			normalizedSelf = filepath.Clean(selfExePath)
		}
	}

	seen := make(map[string]bool)
	result := make([]string, 0, len(args))

	for _, raw := range args {
		candidate := strings.TrimSpace(raw)
		candidate = strings.Trim(candidate, `"'`)
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}

		// Resolve relative paths against workingDir when available.
		resolved := candidate
		if !filepath.IsAbs(resolved) && strings.TrimSpace(workingDir) != "" {
			resolved = filepath.Join(workingDir, resolved)
		}
		if abs, err := filepath.Abs(filepath.Clean(resolved)); err == nil {
			resolved = abs
		} else {
			resolved = filepath.Clean(resolved)
		}

		info, err := os.Stat(resolved)
		if strings.HasPrefix(candidate, "-") && (err != nil || info.IsDir()) {
			continue
		}
		if err != nil || info.IsDir() {
			continue
		}

		if normalizedSelf != "" && pathsEqual(resolved, normalizedSelf) {
			continue
		}

		key := pathKey(resolved)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, resolved)
	}

	return result
}

func pathsEqual(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

func pathKey(p string) string {
	if runtime.GOOS == "windows" {
		return strings.ToLower(p)
	}
	return p
}
