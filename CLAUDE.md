# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

It is the authoritative record of this project's conventions and constraints. `docs/design.md`
is authoritative for *behavior* — the functional requirements and the org → Markdown mapping
tables. `openspec/config.yaml` deliberately restates neither: it points here, and carries only
the per-artifact rules that govern proposals, specs, designs, and task lists. Change a
convention here, not in two places.

## Project

`org2md` is a Go CLI that converts an Emacs org-mode file into a Markdown file suitable
for an Obsidian vault. If `MDNOTES_INBOX` is set in the environment, the tool offers to move
the generated Markdown into `$MDNOTES_INBOX`, which names the destination directory itself.

## Repository state

Released at v1.2.0. Every step of the v1 plan in `docs/design.md` is done, and one change
has shipped on top of it: `$MDNOTES_DIR/Inbox` became `$MDNOTES_INBOX`, which must name an
existing directory. `cmd/org2md` parses an org file, renders it, and places the result;
`internal/org` parses, `internal/md` renders, `internal/vault` decides where the output goes
and offers the `$MDNOTES_INBOX` move.

Further work is planned in OpenSpec, not in `docs/design.md` — see **Planning** below. The
open questions at the end of `docs/design.md` are the backlog those changes are drawn from.
What has shipped since the v1 plan is in `openspec/changes/archive/`, one directory per
change, each holding the proposal and design that argued for it. Read the relevant one before
reopening a decision it settled.

Run `make golden` to rewrite the golden files in `internal/md/testdata` and
`cmd/org2md/testdata` after an intended rendering change, and read the resulting diff before
committing it. The `cmd` fixture is produced by the built binary, so `make golden` runs a
build; `go test -short` skips the end-to-end tests that need one.

`docs/design.md` is the v1 design record: numbered functional requirements (FR-1 … FR-7),
the org → Markdown mapping tables, the package design, and the completed implementation plan.
Read it before writing code — it is still the reference for how the tool behaves today, and
it is kept that way rather than left frozen. FR-2 has been amended twice by the
`$MDNOTES_INBOX` change, so it describes current behavior, not what was originally written
there. Do not add plan steps to it; update it when a change settles one of its open questions
or contradicts a documented decision, rather than letting the code drift from it.

Settled decisions:

- Module path `github.com/go-spass/org2md`
- `cmd/org2md/` for CLI wiring, `internal/org` (parser), `internal/md` (renderer),
  `internal/vault` (output placement and the `MDNOTES_INBOX` move)
- `spf13/cobra` for the CLI
- Conversion scope: core org syntax + YAML frontmatter + task metadata. Set in v1 and
  unchanged since — v1.2.0 altered output placement only, and no org construct renders
  differently.
- Output: explicit `-o` wins; otherwise write beside the source and offer the move.
  The target vault does not use daily notes, so dates never render as links.
- One file per invocation. No bulk, batch, or directory mode, and no multi-file subcommand.
  org2md is used surgically: a rough note captured in org-mode is converted when it is ready
  for a more formal place in the vault. Issue #14 closed on this, and it retires `docs/design.md`
  open question 3, which had called a directory mode "the obvious first subcommand".
- Versioning stays in the 1.x series, and a breaking CLI change does not bump the major.
  org2md is a command, not a library: everything importable lives under `internal/`, which
  Go forbids anyone outside this module from importing, so there is no public Go API for a
  major version to describe. Go's major-version rule (Semantic Import Versioning) would
  require the module path to become `github.com/go-spass/org2md/v2` and every internal import
  to be rewritten — and until it is, the Go proxy rejects a v2 tag outright and
  `go install ...@latest` silently resolves to the newest 1.x instead. Paying that to signal
  a renamed environment variable is the wrong trade for a tool installed with `go install`.
  A breaking CLI change is announced in the release notes and carries `feat!:` with a
  `BREAKING CHANGE:` footer in the commit — Conventional Commits' `!` marks the break, it
  does not oblige a major bump. Revisit only if org2md ever exports a package outside
  `internal/`.

The open items are listed at the end of `docs/design.md` — raise them rather than deciding
them silently.

## Workflow

**Test-driven, without exception.** Write the failing test first, run it, and report that it
failed before writing a line of production code. Then the minimum code to make it pass, then
refactor. No production code lands without a test that drove it. When fixing a bug,
reproduce it as a failing test before touching the fix. If a piece of production code cannot
be driven by a test, stop and say so rather than writing it.

**One branch and one PR per task.** Each task in a change's `tasks.md` gets its own branch,
cut fresh from `main`, and its own pull request. Never carry a second task onto a branch that
already holds one. Because a task is its own PR, the whole red-green-refactor cycle happens
inside it — the failing test and the code that satisfies it land together, never split across
two PRs. Tasks are sized to be independently shippable, so `main` builds and passes its tests
at every merge. The change itself is archived once its tasks have merged.

**Feature branches only.** Never commit to `main`. Branch first (`feat/...`, `fix/...`,
`refactor/...`, `test/...`, `docs/...` matching the conventional-commit type), make the
change there, and open a PR. If work has already started on `main`, move it to a branch
before committing.

Do not commit unless asked, and never pass `--no-verify`.

## Planning

Planning goes through OpenSpec (`openspec` CLI, `spec-driven` schema). `openspec/specs/`
holds the capability specs — what the tool must do; `openspec/changes/` holds in-flight
proposals; `openspec/config.yaml` carries the project context and the per-artifact rules
that every proposal, spec, design, and task list must follow.

The loop: `openspec propose` (or `/opsx:propose`) to turn a request into proposal, specs,
design, and tasks → `openspec apply` to implement them red-green → `openspec archive` when
the PR merges. `openspec explore` first when the shape of the work is still unclear.
`openspec list --specs` is the capability inventory; read it before naming a new capability
so deltas modify existing specs rather than minting near-duplicates.

Proposals are planning only. Do not edit code in the same step that writes the artifacts —
wait for the apply step.

## Commands

Go 1.27 (`go version go1.27.3 darwin/arm64`).

`make` is the interface for building and testing — always use the Makefile targets rather
than calling `go` directly:

```sh
make all            # vet, test, then build
make build          # build the org2md binary
make install        # install the binary
make test           # go test ./...
make test-race      # go test -race ./...
make golden         # rewrite the golden files, then read the diff
make cover          # go test -cover ./...
make vet            # go vet ./...
make fmt            # gofmt -w .
make check          # vet, then test
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
- `Render` returns `(string, []string)`: the Markdown and the warnings, in document order.
  Anything org can say that Markdown cannot is reported rather than dropped in silence, with
  `:LOGBOOK:` the one deliberate exception.
- Tests are table-driven snippets per construct, plus `testdata/` golden files for whole
  documents. `internal/vault` uses `t.TempDir()` and takes the inbox as a `Placer` field
  rather than from the environment, with the prompt reading from an injected `io.Reader` so
  no test needs a TTY. `cmd/org2md`, which does read the environment, unsets the inbox
  variables in `TestMain`: a developer's own shell has one exported, and a test must never
  reach a real vault.
