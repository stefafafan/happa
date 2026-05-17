package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/stefafafan/happa/internal/check"
)

const usage = `Usage:
  happa [flags] [package ...]

Checks pnpm projects for npm packages and reports direct installed versions and
resolved lockfile versions.

Examples:
  happa react vite
  printf 'react\nvite\n' | happa
  find ~/src -name .git -type d -prune | sed 's#/.git$##' | happa --repo - react
  printf 'repo-a\treact\nrepo-b\tvite\n' | happa --stdin pairs

Flags:
`

type repeated []string

func (r *repeated) String() string {
	return strings.Join(*r, ",")
}

func (r *repeated) Set(value string) error {
	*r = append(*r, value)
	return nil
}

type options struct {
	format string
	stdin  string
	header bool
	repos  repeated
	pkgs   repeated
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	opts, positional, err := parseArgs(args, stderr)
	if err != nil {
		return err
	}

	requests, err := buildRequests(opts, positional, stdin)
	if err != nil {
		return err
	}
	if len(requests) == 0 {
		return errors.New("no packages to check")
	}

	results := check.Requests(requests)
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Repo != results[j].Repo {
			return results[i].Repo < results[j].Repo
		}
		return results[i].Package < results[j].Package
	})

	switch opts.format {
	case "tsv":
		writeTSV(stdout, results, opts.header)
	case "json":
		writeJSON(stdout, results)
	default:
		return fmt.Errorf("unsupported format %q; expected tsv or json", opts.format)
	}

	return nil
}

func parseArgs(args []string, stderr io.Writer) (options, []string, error) {
	opts := options{format: "tsv", stdin: "packages"}

	fs := flag.NewFlagSet("happa", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprint(stderr, usage)
		fs.PrintDefaults()
	}
	fs.Var(&opts.repos, "repo", "Repository path to inspect. Repeatable. Use '-' to read newline-delimited repositories from stdin.")
	fs.Var(&opts.repos, "r", "Repository path to inspect. Repeatable. Use '-' to read newline-delimited repositories from stdin.")
	fs.Var(&opts.pkgs, "package", "Package name to check. Repeatable. Use '-' to read newline-delimited packages from stdin.")
	fs.Var(&opts.pkgs, "p", "Package name to check. Repeatable. Use '-' to read newline-delimited packages from stdin.")
	fs.StringVar(&opts.format, "format", opts.format, "Output format: tsv or json.")
	fs.StringVar(&opts.stdin, "stdin", opts.stdin, "How to interpret implicit stdin: packages, repos, or pairs.")
	fs.BoolVar(&opts.header, "header", opts.header, "Print a header row for tabular output.")

	if err := fs.Parse(args); err != nil {
		return opts, nil, err
	}
	return opts, fs.Args(), nil
}

func buildRequests(opts options, positional []string, stdin io.Reader) ([]check.Request, error) {
	if opts.stdin != "packages" && opts.stdin != "repos" && opts.stdin != "pairs" {
		return nil, fmt.Errorf("unsupported --stdin %q; expected packages, repos, or pairs", opts.stdin)
	}

	repos := withoutSentinel(opts.repos)
	packages := append(withoutSentinel(opts.pkgs), positional...)

	readRepos := contains(opts.repos, "-")
	readPackages := contains(opts.pkgs, "-")

	if opts.stdin == "pairs" {
		if readRepos || readPackages {
			return nil, errors.New("--stdin pairs cannot be combined with --repo - or --package -")
		}
		return readPairRequests(stdin)
	}

	if opts.stdin == "repos" {
		readRepos = true
	}
	if opts.stdin == "packages" && len(packages) == 0 && !readRepos {
		readPackages = true
	}

	if readRepos && readPackages {
		return nil, errors.New("stdin cannot be used for both repositories and packages; use --stdin pairs for repo/package rows")
	}

	if readRepos {
		read, err := readLines(stdin)
		if err != nil {
			return nil, err
		}
		repos = append(repos, read...)
	}
	if readPackages {
		read, err := readLines(stdin)
		if err != nil {
			return nil, err
		}
		packages = append(packages, read...)
	}

	if len(repos) == 0 {
		repos = []string{"."}
	}

	return crossProduct(repos, packages), nil
}

func readPairRequests(r io.Reader) ([]check.Request, error) {
	var requests []check.Request
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid pair row %q; expected: <repo><space><package>", line)
		}
		requests = append(requests, check.Request{Repo: fields[0], Package: fields[1]})
	}
	return requests, scanner.Err()
}

func readLines(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, strings.Fields(line)...)
	}
	return lines, scanner.Err()
}

func crossProduct(repos, packages []string) []check.Request {
	var requests []check.Request
	for _, repo := range unique(repos) {
		for _, pkg := range unique(packages) {
			requests = append(requests, check.Request{Repo: repo, Package: pkg})
		}
	}
	return requests
}

func withoutSentinel(values []string) []string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if value != "-" {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func unique(values []string) []string {
	seen := make(map[string]bool, len(values))
	var out []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func writeTSV(w io.Writer, results []check.Result, header bool) {
	if header {
		fmt.Fprintln(w, "repo\tpackage\tsource\tversion\terror")
	}
	for _, result := range results {
		if result.Status == "missing" {
			continue
		}
		wroteVersion := false
		for _, version := range result.InstalledVersions {
			fmt.Fprintf(w, "%s\t%s\tinstalled\t%s\t\n", result.Repo, result.Package, version)
			wroteVersion = true
		}
		for _, version := range result.ResolvedVersions {
			fmt.Fprintf(w, "%s\t%s\tresolved\t%s\t\n", result.Repo, result.Package, version)
			wroteVersion = true
		}
		if !wroteVersion {
			fmt.Fprintf(w, "%s\t%s\t%s\t\t%s\n", result.Repo, result.Package, result.Status, result.Error)
		}
	}
}

func writeJSON(w io.Writer, results []check.Result) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(results)
}
