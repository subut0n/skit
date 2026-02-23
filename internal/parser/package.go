package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Script represents a single script entry from package.json.
type Script struct {
	Name        string // e.g. "test", "build", "dev"
	Command     string // e.g. "vitest run", "next build"
	Description string // from x-skit or empty
	Group       string // prefix before ":" (e.g. "test" for "test:watch")
}

// WorkspaceInfo represents a sub-project in a monorepo.
type WorkspaceInfo struct {
	Name    string // package name from package.json (e.g. "@acme/web")
	Path    string // relative directory path (e.g. "apps/web")
	PkgPath string // absolute path to the package.json
}

// packageJSON is the minimal structure we need from package.json.
type packageJSON struct {
	Name       string            `json:"name"`
	Scripts    map[string]string `json:"scripts"`
	XSkit      map[string]string `json:"x-skit"`
	Workspaces workspacesField   `json:"-"`
}

// workspacesField handles both "workspaces": ["a/*"] and "workspaces": {"packages": ["a/*"]}.
type workspacesField []string

func (w *workspacesField) UnmarshalJSON(data []byte) error {
	// Try array first: ["apps/*", "packages/*"]
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*w = arr
		return nil
	}
	// Try object: {"packages": ["apps/*", "packages/*"]}
	var obj struct {
		Packages []string `json:"packages"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		*w = obj.Packages
		return nil
	}
	return nil
}

// fullPackageJSON includes the workspaces field for monorepo detection.
type fullPackageJSON struct {
	Name       string            `json:"name"`
	Scripts    map[string]string `json:"scripts"`
	XSkit      map[string]string `json:"x-skit"`
	Workspaces workspacesField   `json:"workspaces"`
}

// Parse reads a package.json file and returns its scripts sorted by group then name.
func Parse(path string) ([]Script, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	if len(pkg.Scripts) == 0 {
		return nil, nil
	}

	// First pass: collect all group prefixes from scripts with ":"
	groups := make(map[string]bool)
	for name := range pkg.Scripts {
		if idx := strings.IndexByte(name, ':'); idx > 0 {
			groups[name[:idx]] = true
		}
	}

	scripts := make([]Script, 0, len(pkg.Scripts))
	for name, cmd := range pkg.Scripts {
		s := Script{
			Name:    name,
			Command: cmd,
		}

		// Extract group: either from ":" prefix or by matching an existing group
		if idx := strings.IndexByte(name, ':'); idx > 0 {
			s.Group = name[:idx]
		} else if groups[name] {
			// Script name matches a group prefix (e.g. "test" when "test:watch" exists)
			s.Group = name
		}

		// Description from x-skit field
		if desc, ok := pkg.XSkit[name]; ok {
			s.Description = desc
		}

		scripts = append(scripts, s)
	}

	// Sort: ungrouped first, then by group, then by name within each group
	sort.Slice(scripts, func(i, j int) bool {
		gi, gj := scripts[i].Group, scripts[j].Group
		if gi == "" && gj != "" {
			return true
		}
		if gi != "" && gj == "" {
			return false
		}
		if gi != gj {
			return gi < gj
		}
		return scripts[i].Name < scripts[j].Name
	})

	return scripts, nil
}

// ParseName reads just the "name" field from a package.json.
func ParseName(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var pkg struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return ""
	}
	return pkg.Name
}

// FindPackageJSON searches for package.json starting from dir and walking up parent directories.
func FindPackageJSON(dir string) string {
	for {
		path := dir + "/package.json"
		if _, err := os.Stat(path); err == nil {
			return path
		}

		parent := dir[:strings.LastIndex(dir, "/")]
		if parent == dir || parent == "" {
			break
		}
		dir = parent
	}
	return ""
}

// FindRootPackageJSON walks up from dir to find the topmost package.json.
func FindRootPackageJSON(dir string) string {
	var root string
	for {
		path := dir + "/package.json"
		if _, err := os.Stat(path); err == nil {
			root = path
		}

		parent := dir[:strings.LastIndex(dir, "/")]
		if parent == dir || parent == "" {
			break
		}
		dir = parent
	}
	return root
}

// ParseWorkspaces reads workspace patterns from a root package.json (npm/yarn/bun)
// or from pnpm-workspace.yaml, then resolves all sub-project package.json files.
func ParseWorkspaces(rootPkgPath string) []WorkspaceInfo {
	rootDir := filepath.Dir(rootPkgPath)
	patterns := readWorkspacePatterns(rootPkgPath, rootDir)
	if len(patterns) == 0 {
		return nil
	}

	var workspaces []WorkspaceInfo
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		// Resolve the glob pattern relative to root
		absPattern := filepath.Join(rootDir, pattern)

		matches, err := filepath.Glob(absPattern)
		if err != nil {
			continue
		}

		for _, match := range matches {
			pkgPath := filepath.Join(match, "package.json")
			if _, err := os.Stat(pkgPath); err != nil {
				continue
			}

			// Deduplicate
			abs, _ := filepath.Abs(pkgPath)
			if seen[abs] {
				continue
			}
			seen[abs] = true

			relPath, _ := filepath.Rel(rootDir, match)
			name := ParseName(pkgPath)
			if name == "" {
				name = relPath
			}

			workspaces = append(workspaces, WorkspaceInfo{
				Name:    name,
				Path:    relPath,
				PkgPath: pkgPath,
			})
		}
	}

	sort.Slice(workspaces, func(i, j int) bool {
		return workspaces[i].Path < workspaces[j].Path
	})

	return workspaces
}

// readWorkspacePatterns extracts workspace glob patterns from package.json or pnpm-workspace.yaml.
func readWorkspacePatterns(pkgPath, rootDir string) []string {
	// Try package.json workspaces field (npm/yarn/bun)
	data, err := os.ReadFile(pkgPath)
	if err == nil {
		var pkg fullPackageJSON
		if err := json.Unmarshal(data, &pkg); err == nil && len(pkg.Workspaces) > 0 {
			return pkg.Workspaces
		}
	}

	// Try pnpm-workspace.yaml
	pnpmPath := filepath.Join(rootDir, "pnpm-workspace.yaml")
	pnpmData, err := os.ReadFile(pnpmPath)
	if err != nil {
		return nil
	}
	return parsePnpmWorkspaceYAML(string(pnpmData))
}

// parsePnpmWorkspaceYAML does a simple line-based parse of pnpm-workspace.yaml.
func parsePnpmWorkspaceYAML(content string) []string {
	var patterns []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") {
			pattern := strings.TrimPrefix(line, "- ")
			pattern = strings.Trim(pattern, "'\"")
			if pattern != "" {
				patterns = append(patterns, pattern)
			}
		}
	}
	return patterns
}
