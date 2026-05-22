# happa - a simple CLI tool for listing installed / resolved npm packages

`happa` checks and prints the installed / resolved versions of specified packages, without needing to run `pnpm list` or `pnpm why` in each repository.

It is designed for easy list up of packages across many repositories at once, and to be used alongside other commands such as `grep`.

The term `happa` comes from the Japanese word for leaf. The name is a play on the leaves of a dependency tree.

> [!IMPORTANT]
> This tool only supports pnpm at the moment.

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

Check repositories with `pnpm-workspace.yaml` from a parent directory:

```sh
find . -mindepth 2 -maxdepth 2 -name pnpm-workspace.yaml -print \
  | sed 's#/pnpm-workspace.yaml$##' \
  | happa --repo - react
```

Default output is TSV. If the package is not found, it is omitted from the output.

```text
happa	react	installed	19.1.0
happa	react	resolved	19.1.0
```

Using alongside `grep`, happa makes it easy to check for specific versions.

```sh
happa react | grep '19.1.0'
```

JSON is also available:

```sh
happa --format json react
```

