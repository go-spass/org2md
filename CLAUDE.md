# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`org2md` is a Go CLI that converts an Emacs org-mode file into a Markdown file suitable
for an Obsidian vault. If `MDNOTES_DIR` is set in the environment, the tool offers to move
the generated Markdown into `$MDNOTES_DIR/Inbox`.

## Repository state

Pre-implementation. The repo holds `README.md`, `LICENSE`, `.gitignore`, `docs/design.md`,
and this file — there is no `go.mod` and no Go source yet.

`docs/design.md` is the spec: numbered functional requirements (FR-1 … FR-7), the org →
Markdown mapping tables, the package design, and a step-by-step implementation plan. Read it
before writing code, and update it when a decision changes rather than letting the code
drift from it.

Settled decisions:

- Module path `github.com/go-spass/org2md`
- `cmd/org2md/` for CLI wiring, `internal/org` (parser), `internal/md` (renderer),
  `internal/vault` (output placement and the `MDNOTES_DIR` move)
- `spf13/cobra` for the CLI
- v1 scope: core org syntax + YAML frontmatter + task metadata
- Output: explicit `-o` wins; otherwise write beside the source and offer the move

The open items are listed at the end of `docs/design.md` — raise them rather than deciding
them silently.

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

- Conversion is pure: `io.Reader` in, string out, no filesystem and no environment access.
  Anything interactive or environment-dependent belongs in `internal/vault` or `main`.
- Parsing is line-oriented in two phases (blocks, then inlines). Inline markup must never be
  applied inside source/example/verbatim content — a bare `*` in a shell snippet has to
  survive. This is why the tool parses instead of running string substitutions.
- `internal/org` owns org syntax; `internal/md` owns Obsidian conventions. A construct with
  no Markdown equivalent is a rendering decision, not a parsing one.
- Unrecognized org syntax passes through verbatim with a warning; it never fails the run.
- Tests are table-driven snippets per construct, plus `testdata/` golden files for whole
  documents. `internal/vault` uses `t.TempDir()` and `t.Setenv`, with the prompt reading from
  an injected `io.Reader` so no test needs a TTY.
