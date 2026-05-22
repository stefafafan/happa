# happa

`happa` checks and prints the installed / resolved versions of specified packages, without needing to run `pnpm list` or `pnpm why` in each repository.

It is designed for easy list up of packages across many repositories at once, and to be used alongside other commands such as `grep`.

The term `happa` comes from the Japanese word for leaf. The name is a play on the leaves of a dependency tree.

## Install

```sh
go install github.com/stefafafan/happa/cmd/happa@latest
```

## Usage

Check the current repository for specified packages (in this case, `react` and `vite`):

```sh
happa react vite
```

You can also pipe package names from other commands, such as `find` or `printf`:

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

Default output is TSV. If the package is not found, it is ommitted from the output.

```text
happa	react	installed	19.1.0
happa	react	resolved	19.1.0
```

Using alongside `grep`, happa makes it easy to check for specific versions.

```sh
happa react | grep '19.1.0'
```

Check multiple repositories for a specific version:

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

- `installed`: found in a pnpm `node_modules` tree by reading the installed package's own `package.json`.
- `resolved`: found in `pnpm-lock.yaml` under `packages` or `snapshots`, which includes transitive dependencies.
