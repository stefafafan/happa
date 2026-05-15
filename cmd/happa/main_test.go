package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stefafafan/happa/internal/check"
)

func TestRunReadsPackagesFromStdin(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := run([]string{"--format", "json"}, strings.NewReader("react\nvite\n"), &stdout, &stderr)
	if err != nil {
		t.Fatalf("run returned error: %v; stderr=%s", err, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, `"package": "react"`) || !strings.Contains(output, `"package": "vite"`) {
		t.Fatalf("output did not contain stdin packages: %s", output)
	}
}

func TestBuildRequestsReadsReposFromStdin(t *testing.T) {
	requests, err := buildRequests(
		options{stdin: "repos", pkgs: repeated{"react"}},
		nil,
		strings.NewReader("repo-a\nrepo-b\n"),
	)
	if err != nil {
		t.Fatal(err)
	}

	want := []RequestLike{
		{Repo: "repo-a", Package: "react"},
		{Repo: "repo-b", Package: "react"},
	}
	assertRequests(t, requests, want)
}

func TestBuildRequestsReadsPairRowsFromStdin(t *testing.T) {
	requests, err := buildRequests(options{stdin: "pairs"}, nil, strings.NewReader("repo-a react\nrepo-b vite\n"))
	if err != nil {
		t.Fatal(err)
	}

	want := []RequestLike{
		{Repo: "repo-a", Package: "react"},
		{Repo: "repo-b", Package: "vite"},
	}
	assertRequests(t, requests, want)
}

func TestWriteTSVOmitsHeaderByDefault(t *testing.T) {
	var output bytes.Buffer

	writeTSV(&output, []check.Result{{
		Repo:              ".",
		Package:           "react",
		Status:            "installed+resolved",
		InstalledVersions: []string{"19.1.0"},
		ResolvedVersions:  []string{"19.1.0"},
	}}, false)

	want := ".\treact\tinstalled\t19.1.0\t\n.\treact\tresolved\t19.1.0\t\n"
	if output.String() != want {
		t.Fatalf("writeTSV() = %q, want %q", output.String(), want)
	}
}

func TestWriteTSVCanIncludeHeader(t *testing.T) {
	var output bytes.Buffer

	writeTSV(&output, []check.Result{{
		Repo:    ".",
		Package: "react",
		Status:  "missing",
	}}, true)

	want := "repo\tpackage\tsource\tversion\terror\n.\treact\tmissing\t\t\n"
	if output.String() != want {
		t.Fatalf("writeTSV() = %q, want %q", output.String(), want)
	}
}

type RequestLike struct {
	Repo    string
	Package string
}

func assertRequests(t *testing.T, got []check.Request, want []RequestLike) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d requests, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].Repo != want[i].Repo || got[i].Package != want[i].Package {
			t.Fatalf("request %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}
