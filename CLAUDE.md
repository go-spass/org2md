# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`org2md` is a Go CLI that converts an Emacs org-mode file into a Markdown file suitable
for an Obsidian vault. If `MDNOTES_DIR` is set in the environment, the tool offers to move
the generated Markdown into `$MDNOTES_DIR/Inbox`.

## Repository state

Pre-implementation. The repo currently holds only `README.md`, `LICENSE`, `.gitignore`, and
this file — there is no `go.mod` and no Go source yet. Treat the README as the spec.

Before writing code, confirm with the user:
- the module path (the remote is `github.com/go-spass/org2md`)
- the layout (single-package `main` at the root vs. `cmd/org2md` + internal packages)
- the CLI framework (stdlib `flag` vs. Cobra/urfave-cli)

Do not scaffold these choices unilaterally.

## Workflow

**Test-driven.** Write the failing test first, watch it fail, then write the minimum code to
make it pass, then refactor. No production code lands without a test that drove it. When
fixing a bug, reproduce it as a failing test before touching the fix.

**Feature branches only.** Never commit to `main`. Branch first (`feat/...`, `fix/...`,
`refactor/...`, `test/...`, `docs/...` matching the conventional-commit type), make the
change there, and open a PR. If work has already started on `main`, move it to a branch
before committing.

## Commands

Go 1.27 (`go version go1.27.3 darwin/arm64`).

`make` is the interface for building and testing — always use the Makefile targets rather
than calling `go` directly. The Makefile does not exist yet; creating it is part of the
initial scaffolding, and it should wrap at least:

```sh
make build          # build the org2md binary
make test           # go test ./...
make test-race      # go test -race ./...
make cover          # go test -cover ./...
make vet            # go vet ./...
make fmt            # gofmt -w .
make clean
```

For a single test during a red-green cycle, `go test -run TestName ./path/to/pkg` directly
is fine; everything else goes through `make`. If a target you need is missing, add it to the
Makefile instead of working around it.

## Design notes

- The org→Markdown conversion and the file-move/Obsidian step are separate concerns; keep
  conversion pure (string/reader in, string/writer out) so it is testable without touching
  the filesystem or `MDNOTES_DIR`.
- The `MDNOTES_DIR` move is interactive ("offer to move"), so it must be skippable — keep
  it out of the conversion path and behind a flag for non-interactive use.
- Table-driven tests with org input and expected Markdown output fit this problem well;
  `testdata/` golden files are the natural place for larger fixtures.
