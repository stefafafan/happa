# happa

`happa` checks pnpm projects for npm packages and prints whether each package is
installed in `node_modules`, resolved in the lockfile dependency graph, or
missing.

It is designed for supply-chain incident checks where you need to ask "which of
these repositories resolve this package, and which version?" across many local
git repositories.

## Install

```sh
go install ./cmd/happa
```

## Usage

Check the current repository:

```sh
happa react vite
```

Pipe package names:

```sh
printf 'react\nvite\n' | happa
```

Pipe repositories while passing packages as arguments:

```sh
find ~/src -name .git -type d -prune | sed 's#/.git$##' | happa --repo - react
```

Pipe explicit repository/package pairs:

```sh
printf 'repo-a\treact\nrepo-b\tvite\n' | happa --stdin pairs
```

Default output is TSV:

```text
happa	react	installed	19.1.0	
happa	react	resolved	19.1.0	
```

That keeps versions as standalone fields, so compromised versions are easy to
filter:

```sh
happa react | grep '19.1.0'
```

Check multiple repositories for a compromised package version:

```sh
happa --repo ~/src/repo-a --repo ~/src/repo-b react | grep -F '19.1.0'
```

Add a TSV header row when needed:

```sh
happa --header react
```

JSON is also available:

```sh
happa --format json react
```

## Meaning

- `installed`: found in a pnpm `node_modules` tree by reading the installed
  package's own `package.json`.
- `resolved`: found in `pnpm-lock.yaml` under `packages` or `snapshots`, which
  includes transitive dependencies.
- `installed+resolved`: found in both places.
- `missing`: not found.
- `error`: repository could not be inspected.

Only modern pnpm lockfiles are supported in this version. The parser targets the
`pnpm-lock.yaml` structure used by pnpm 10/11 projects: `importers`,
`packages`, and `snapshots` with package keys such as `react@19.1.0` and
`@scope/pkg@2.1.0`.
