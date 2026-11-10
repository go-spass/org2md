# Design

## Context

See `proposal.md` — Why. Three facts about the current code shape the approach.

`Placer` carries a design note that this change should honor rather than erode:

> Everything it needs from the environment is a field, so its behaviour is decided by the
> caller rather than discovered here.

`NotesDir` is read in exactly one place, `inboxPath` at `place.go:78`, which joins
`const inbox = "Inbox"` onto it. The vault root has no other purpose in the codebase, so the
subdirectory join is org2md's own opinion layered on a path that otherwise does nothing.

`Place` writes beside the input at `place.go:56` *before* it looks at the inbox at `:60`.
That ordering is why today's failure is silent, and it is what has to change for an error to
mean "nothing was written".

## Goals / Non-Goals

**Goals:**

- One environment variable that names the destination directory, with no derived path.
- A misconfigured inbox produces a message, not filesystem structure.
- An error that leaves the working tree exactly as it was.
- Keep `Placer`'s "environment arrives as fields" property intact.

**Non-Goals:**

- Beyond the proposal's out-of-scope list, this design does not restructure `Place`'s
  control flow past moving the validation ahead of the first write, and does not change how
  the cross-filesystem copy fallback in `move` works.

## Decisions

**The variable names the destination, and the constant is deleted.**
`inboxPath` becomes `filepath.Join(p.InboxDir, filepath.Base(written))`. Alternatives: keep
`$MDNOTES_DIR` and add `$MDNOTES_INBOX` as the *subdirectory name* to join onto it (issue
#13's original proposal) — rejected because it leaves two variables to keep in sync and keeps
a derived path that can be wrong; or redefine `$MDNOTES_DIR` to point at the capture folder —
rejected because an existing setup would silently start filing to the vault root, which is a
worse failure than the one being fixed. The rename makes the old configuration inert rather
than wrong.

**`NotesDir` becomes `InboxDir`.** The field is renamed rather than reinterpreted, so no
caller can keep passing a vault root into a field that now means something else. The doc
comment changes from `$MDNOTES_DIR, empty when unset` to `$MDNOTES_INBOX, empty when unset`.

**Validation lives in `internal/vault`; reading the variable lives in `cmd/org2md`.**
`os.Getenv` stays in `root.go`, where environment access already belongs, and the directory
check goes next to the other filesystem operations in `place.go`. Alternative: stat the
directory in `main` before constructing the `Placer` — rejected because it splits one rule
across two packages and leaves `vault` willing to move into a path it never checked, so the
package's own tests could not cover the rule.

**The check runs at the top of `Place`, gated on a move actually being in play.**
Only when `output == ""`, `InboxDir != ""`, and `!NoMove` — so `-o`, `-o -`, and `--no-move`
runs never stat the inbox and are unaffected by a bad value. Alternatives: check inside
`inboxPath` — rejected because by then the beside-the-input write has already happened, so
the error would contradict the `warning-reporting` requirement that fatal conditions write
nothing; or check inside `move`, after confirmation — rejected because it asks a move
question that cannot be honored and then fails anyway.

**`os.MkdirAll` is removed from `move` rather than made conditional.** Once the inbox is
validated up front, the only directory `move` could create is one that was just confirmed to
exist. Keeping the call would mean a second, unreachable policy about inventing directories.

**Relative values are accepted as-is.** `filepath.Join` resolves them against the working
directory. This is surprising, but a relative value that does not resolve to an existing
directory now errors immediately, so it is self-correcting rather than silent. Rejecting
non-absolute paths would add a rule whose only remaining job is to editorialize.

## Risks / Trade-offs

- **A misconfigured inbox now loses the conversion output entirely**, where before the
  Markdown at least landed beside the input → the error names the configured path, so the fix
  is one shell line and a rerun; conversion is fast, and `--no-move`, `-o`, or `-o -` all
  bypass the check if the output is wanted immediately.
- **Breaking for anyone with `$MDNOTES_DIR` set** → the repo is private and both releases are
  source-only, so the blast radius is one shell profile; ship as v2.0.0 with the README and
  `--help` updated. Issue #15 closes this window.
- **A genuinely fresh vault no longer gets its inbox created for it** → one `mkdir -p`, and
  the error message names the path that is missing. Trading a convenience that fires once for
  a failure mode that fires on every typo is the right way round.
- **The `~` quoting trap**: `MDNOTES_INBOX="~/vault/00 - Inbox"` arrives literally, because
  the space forces quoting and `~` does not expand inside quotes → it becomes an error naming
  a path starting with `~`, which is self-diagnosing. No tilde expansion is added.

## Migration Plan

Replace the variable in the shell profile:

```sh
# before
export MDNOTES_DIR="$HOME/Documents/second-brain"
# after
export MDNOTES_INBOX="$HOME/Documents/second-brain/00 - Inbox"
```

Use `$HOME` rather than `~`, since the directory name contains spaces and must be quoted.
Rollback is the previous tag; nothing persists outside the binary, so there is no state to
undo.

## Package ownership

| Package | Change |
| --- | --- |
| `cmd/org2md` | `notesDirEnv` → `MDNOTES_INBOX`, `os.Getenv` into the `InboxDir` field, help text and examples |
| `internal/vault` | `InboxDir` field, deleted `inbox` constant, inbox validation, `MkdirAll` removed |
| `internal/org` | untouched |
| `internal/md` | untouched |

The purity boundary is preserved and in fact tightened: the conversion path never learns that
an inbox exists. `internal/org` and `internal/md` keep taking an `io.Reader` and returning a
string with no filesystem or environment access, the environment is read only in
`cmd/org2md`, and the only new filesystem call is a `Stat` inside `internal/vault`, which is
where the existing writes and moves already live.

This change does not touch the org → Markdown mapping tables in `docs/design.md`: no org
construct renders differently. The edit needed there is to FR-2 (Output location), which
names `$MDNOTES_DIR/Inbox`.

## Open Questions

- Whether a `--inbox <path>` flag should later override the variable for a one-off run. It
  would be additive, needs no change to these specs or tasks, and there is no demand for it
  yet.
