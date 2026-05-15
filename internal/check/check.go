package check

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Request struct {
	Repo    string
	Package string
}

type Result struct {
	Repo              string   `json:"repo"`
	Package           string   `json:"package"`
	Status            string   `json:"status"`
	InstalledVersions []string `json:"installed_versions,omitempty"`
	ResolvedVersions  []string `json:"resolved_versions,omitempty"`
	Error             string   `json:"error,omitempty"`
}

type repoIndex struct {
	Installed map[string]map[string]bool
	Resolved  map[string]map[string]bool
	Err       error
}

func Requests(requests []Request) []Result {
	indexes := map[string]repoIndex{}
	results := make([]Result, 0, len(requests))

	for _, request := range requests {
		repo := filepath.Clean(request.Repo)
		if _, ok := indexes[repo]; !ok {
			indexes[repo] = inspectRepo(repo)
		}
		results = append(results, buildResult(displayRepoName(repo), request.Package, indexes[repo]))
	}

	return results
}

func displayRepoName(repo string) string {
	clean := filepath.Clean(repo)
	if clean == "." || clean == ".." {
		abs, err := filepath.Abs(clean)
		if err == nil {
			return filepath.Base(abs)
		}
	}

	base := filepath.Base(clean)
	if base == "." {
		return clean
	}
	return base
}

func buildResult(repo, pkg string, index repoIndex) Result {
	result := Result{Repo: repo, Package: pkg}
	if index.Err != nil {
		result.Status = "error"
		result.Error = index.Err.Error()
		return result
	}

	result.InstalledVersions = sortedVersions(index.Installed[pkg])
	result.ResolvedVersions = sortedVersions(index.Resolved[pkg])

	switch {
	case len(result.InstalledVersions) > 0 && len(result.ResolvedVersions) > 0:
		result.Status = "installed+resolved"
	case len(result.InstalledVersions) > 0:
		result.Status = "installed"
	case len(result.ResolvedVersions) > 0:
		result.Status = "resolved"
	default:
		result.Status = "missing"
	}
	return result
}

func inspectRepo(repo string) repoIndex {
	info, err := os.Stat(repo)
	if err != nil {
		return repoIndex{Err: err}
	}
	if !info.IsDir() {
		return repoIndex{Err: fmt.Errorf("%s is not a directory", repo)}
	}

	index := repoIndex{
		Installed: map[string]map[string]bool{},
		Resolved:  map[string]map[string]bool{},
	}

	lockPath := filepath.Join(repo, "pnpm-lock.yaml")
	lockBytes, err := os.ReadFile(lockPath)
	hasLock := err == nil
	if err != nil && !os.IsNotExist(err) {
		return repoIndex{Err: err}
	}

	if hasLock || hasPnpmInstall(repo) {
		if err := inspectNodeModules(repo, index.Installed); err != nil {
			return repoIndex{Err: err}
		}
	}

	if hasLock {
		lock := parsePnpmLock(string(lockBytes))
		merge(index.Resolved, lock.Resolved)
	}
	return index
}

func hasPnpmInstall(repo string) bool {
	info, err := os.Stat(filepath.Join(repo, "node_modules", ".pnpm"))
	return err == nil && info.IsDir()
}

func addVersion(target map[string]map[string]bool, name, version string) {
	name = strings.TrimSpace(unquote(name))
	version = normalizeVersion(strings.TrimSpace(unquote(version)))
	if name == "" || version == "" {
		return
	}
	if target[name] == nil {
		target[name] = map[string]bool{}
	}
	target[name][version] = true
}

func merge(target, source map[string]map[string]bool) {
	for name, versions := range source {
		for version := range versions {
			addVersion(target, name, version)
		}
	}
}

func sortedVersions(versions map[string]bool) []string {
	if len(versions) == 0 {
		return nil
	}
	sorted := make([]string, 0, len(versions))
	for version := range versions {
		sorted = append(sorted, version)
	}
	sort.Strings(sorted)
	return sorted
}
