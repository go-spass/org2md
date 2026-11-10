# Proposal

## Why

`internal/vault/place.go` hard-codes `const inbox = "Inbox"` and joins it onto `$MDNOTES_DIR`
unconditionally, so org2md can only file notes into a vault whose capture folder is named
exactly `Inbox`. The vault this tool was built for is PARA-structured and names it
`00 - Inbox`, so accepting the move prompt does not file the note: `move()` calls
`os.MkdirAll`, which silently creates a second, empty `Inbox/` at the vault root and puts the
note there. Nothing errors and nothing warns, so the note lands where the weekly review never
looks (issue #13).

This is live, not hypothetical — `MDNOTES_DIR` is currently set to that vault, which has no
`Inbox/` directory.

## What Changes

- **BREAKING** `$MDNOTES_DIR` is no longer read by org2md. It is a vault root that other tools
  own; org2md has no use for it, and will not read, check, or mention it.
- **BREAKING** `$MDNOTES_INBOX` replaces it, and names the destination directory **directly**.
  There is no subdirectory join: the note is moved to `$MDNOTES_INBOX/<basename>.md`. The
  `inbox` constant is deleted rather than made configurable.
- **BREAKING** A configured inbox that does not exist, or that is not a directory, is an
  error. `os.MkdirAll` is removed from the move path — org2md files into a vault, it does not
  reshape one. A typo in the variable now produces a message instead of an invented directory
  tree.
- The check happens before anything is written, so a bad `$MDNOTES_INBOX` leaves no file
  behind and never asks a move question it cannot honor.
- `$MDNOTES_INBOX` unset or empty means no move is offered at all, exactly as an unset
  `$MDNOTES_DIR` does today. The note stays beside the input.
- Help text, README, and `docs/design.md` FR-2 are updated to the new variable.

Why the rename rather than redefining `$MDNOTES_DIR` to point at the capture folder: an
existing setup would keep working and silently start filing to the vault root, which is a
worse failure than the one being fixed. With a new name, an old setup has `$MDNOTES_INBOX`
unset, so the move is simply never offered and the note stays beside the input where
`--verbose` reports it. Nothing is misfiled.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `output-placement`: the move destination is `$MDNOTES_INBOX` directly rather than
  `$MDNOTES_DIR/Inbox`, and a missing or non-directory inbox becomes an error that writes
  nothing instead of a directory org2md creates.
- `warning-reporting`: the enumeration of fatal conditions gains the unusable inbox, so the
  requirement that fatal conditions write nothing stays complete.

## Impact

Affected code:

- `internal/vault/place.go` — delete `const inbox`, rename the `NotesDir` field to `InboxDir`,
  drop the subdirectory join in `inboxPath`, remove `os.MkdirAll` from `move`, and validate the
  inbox before the beside-the-input write in `Place`.
- `cmd/org2md/root.go` — `notesDirEnv` becomes `MDNOTES_INBOX`, and the `--no-move` flag help,
  the long description, and the examples lose the `/Inbox` suffix.

Affected tests: `internal/vault/place_test.go` and `cmd/org2md/root_test.go` set the old
variable throughout, and `cmd/org2md/e2e_test.go` asserts the `Inbox` subdirectory.
`TestMoveReportsAnUncreatableInbox` exists to prove `MkdirAll` fails usefully, so its premise
is what this change deletes; it becomes a test that a non-directory inbox errors.

Affected docs: `README.md` (4 places), `CLAUDE.md` (3 places), and `docs/design.md` FR-2 and
its package table.

Breaking for anyone with `$MDNOTES_DIR` set, which is why this is worth doing now: the repo is
private and both releases are source-only, so the cost is one line in a shell profile. Issue
#15 (whether this repo goes public) closes that window. Semver puts this at v2.0.0.

## Out of scope

- Reading, checking, or warning about `$MDNOTES_DIR` in any form. No migration shim, no
  deprecation notice, no fallback. The variable is simply not this tool's concern.
- A `--inbox` flag. The destination is set once per machine, not per invocation.
- Requiring `$MDNOTES_INBOX` to be absolute. A relative value resolves against the working
  directory, which is surprising but self-correcting now that a non-existent inbox errors.
- Any flag to restore directory creation. `MkdirAll` is removed, not made optional.
- The semantics of `-o`, `-o -`, `--no-move`, and `--force`, which are unchanged.
- Issue #13's suggestion of keeping `Inbox` as a default. With the variable naming the
  destination directly, a default would mean guessing a directory name, which is the bug.

Two edits this change needs that are not delta artifacts, noted here so they are not lost:
`openspec/specs/output-placement/spec.md`'s `## Purpose` still says "an Obsidian vault's
Inbox" and must be edited in the main spec directly, since a delta for an existing capability
does not carry Purpose; and `docs/design.md` FR-2 is contradicted by this change, which the
project's archive guidance says to update rather than let drift.

This change extends FR-2 (Output location) in `docs/design.md`, and settles the
"Should a missing inbox directory still be created?" question raised in issue #13.
