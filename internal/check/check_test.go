package check

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRequestsReportsInstalledResolvedAndMissing(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, repo, "pnpm-lock.yaml", `
lockfileVersion: '9.0'
importers:
  .:
    dependencies:
      direct:
        specifier: ^1.0.0
        version: 1.2.3
packages:
  direct@1.2.3:
    resolution: {integrity: sha512-test}
  transitive@4.5.6:
    resolution: {integrity: sha512-test}
`)
	writeFile(t, repo, "node_modules/.pnpm/direct@1.2.3/node_modules/direct/package.json", `{
  "name": "direct",
  "version": "1.2.3"
}`)
	writeFile(t, repo, "node_modules/.pnpm/installed-only@9.9.9/node_modules/installed-only/package.json", `{
  "name": "installed-only",
  "version": "9.9.9"
}`)

	results := Requests([]Request{
		{Repo: repo, Package: "direct"},
		{Repo: repo, Package: "transitive"},
		{Repo: repo, Package: "installed-only"},
		{Repo: repo, Package: "missing"},
	})

	byPackage := map[string]Result{}
	for _, result := range results {
		byPackage[result.Package] = result
	}

	if byPackage["direct"].Status != "installed+resolved" {
		t.Fatalf("direct status = %q", byPackage["direct"].Status)
	}
	if got := byPackage["direct"].InstalledVersions; len(got) != 1 || got[0] != "1.2.3" {
		t.Fatalf("direct installed versions = %v", got)
	}
	if byPackage["transitive"].Status != "resolved" {
		t.Fatalf("transitive status = %q", byPackage["transitive"].Status)
	}
	if byPackage["installed-only"].Status != "installed" {
		t.Fatalf("installed-only status = %q", byPackage["installed-only"].Status)
	}
	if byPackage["missing"].Status != "missing" {
		t.Fatalf("missing status = %q", byPackage["missing"].Status)
	}
}

func TestRequestsDoesNotTreatNpmNodeModulesAsPnpmInstall(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, repo, "node_modules/react/package.json", `{
  "name": "react",
  "version": "19.1.0"
}`)

	results := Requests([]Request{{Repo: repo, Package: "react"}})
	if results[0].Status != "missing" {
		t.Fatalf("status = %q, want missing", results[0].Status)
	}
}

func TestRequestsReportsRepoDirectoryName(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "readable-repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}

	results := Requests([]Request{{Repo: repo, Package: "react"}})
	if results[0].Repo != "readable-repo" {
		t.Fatalf("repo = %q, want readable-repo", results[0].Repo)
	}
}

func TestRequestsReportsCurrentDirectoryNameForDotRepo(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "current-repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo)

	results := Requests([]Request{{Repo: ".", Package: "react"}})
	if results[0].Repo != "current-repo" {
		t.Fatalf("repo = %q, want current-repo", results[0].Repo)
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
