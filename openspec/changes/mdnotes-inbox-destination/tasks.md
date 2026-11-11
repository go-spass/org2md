# Tasks

Every tracked task is one branch and one pull request, cut fresh from `main`, carrying its own
failing test first. No task here changes rendering output, so none needs `make golden`.

Design open question (a possible `--inbox` flag) is deferred by design and changes nothing
below.

## 1. Point the move at $MDNOTES_INBOX

- [x] 1.1 Branch `feat/mdnotes-inbox-env`. **BREAKING.** Red first in `internal/vault`:
  retarget `TestPlaceMovesIntoInbox` (`place_test.go:131`) to set an `InboxDir` field and
  expect the note at `<InboxDir>/note.md` with no `Inbox` segment, and add
  `TestPlaceMovesIntoAnInboxWithSpaces` covering a `00 - Inbox` destination. In `cmd/org2md`,
  retarget `TestMovesIntoTheNotesDirectory` (`root_test.go:309`) and
  `TestNoNotesDirectoryMeansNoPrompt` (`root_test.go:354`) to `MDNOTES_INBOX`, add a test that
  a set `MDNOTES_DIR` with `MDNOTES_INBOX` unset offers no prompt and mentions nothing, and
  retarget `TestEndToEndFilesIntoTheInbox` (`e2e_test.go:94`). Then green: delete
  `const inbox` (`place.go:15`), rename `NotesDir` to `InboxDir` with its doc comment,
  simplify `inboxPath` to `filepath.Join(p.InboxDir, filepath.Base(written))`, and change
  `notesDirEnv` to `MDNOTES_INBOX` in `root.go` along with the long description, the examples,
  and the `--no-move` flag help, dropping the `/Inbox` suffix from each. Verify with
  `make check` and that `grep -rn MDNOTES_DIR --include='*.go' .` returns nothing.
- [ ] 1.2 Branch `docs/mdnotes-inbox-env`, immediately after 1.1 merges. Documentation for the
  work 1.1 landed, so no test drives it and no behavior changes: update `README.md` lines 7-8,
  33-34, and the `--no-move` row at 44; `CLAUDE.md` lines 8-9, 15, and 36; and `docs/design.md`
  FR-2's resolution order, the `--no-move` bullet at line 36, and the `internal/vault` row of
  the package table at line 158. Verify by running the README's documented commands as written
  and confirming `grep -rn MDNOTES_DIR README.md CLAUDE.md docs/design.md` returns nothing.

## 2. Require the inbox to exist

- [ ] 2.1 Branch `feat/require-existing-inbox`. **BREAKING.** Red first in `internal/vault`:
  add `TestPlaceRefusesAMissingInbox` and `TestPlaceRefusesANonDirectoryInbox` expecting a
  non-nil error naming the configured path and no directory created;
  `TestPlaceLeavesNothingBehindWhenTheInboxIsMissing` asserting the beside-the-input Markdown
  does not exist afterwards; and `TestPlaceDoesNotAskWhenTheInboxIsUnusable` asserting the
  prompt writer stays empty. Retarget `TestMoveReportsAnUncreatableInbox`
  (`place_test.go:290`), whose premise — that `MkdirAll` fails usefully — this task deletes,
  to assert that a non-directory inbox errors. Add a case that `-o`, `-o -`, and `--no-move`
  runs succeed with an unusable inbox, proving the check is gated on a move being in play.
  Then green: validate `InboxDir` with `os.Stat` plus `IsDir` at the top of `Place`, before the
  write at `place.go:56`, only when `output == ""`, `InboxDir != ""`, and `!NoMove`. Refactor
  step, no behavior change: delete `os.MkdirAll` from `move` (`place.go:121`), which the
  validation makes unreachable. Verify with `make check` and `make test-race`.
- [ ] 2.2 Branch `docs/require-existing-inbox`, immediately after 2.1 merges. Documentation for
  the work 2.1 landed, so no test drives it and no behavior changes: say in `README.md` and in
  `docs/design.md` FR-2 that `$MDNOTES_INBOX` must name an existing directory and that org2md
  does not create it, recommending `$HOME` over `~` because the directory name contains
  spaces. Edit `openspec/specs/output-placement/spec.md`'s `## Purpose` line directly to drop
  "an Obsidian vault's Inbox" — a delta for an existing capability does not carry Purpose, so
  this cannot ride in the change's spec files. Verify with
  `openspec validate --specs --strict`.

## 3. Integration

- [ ] 3.1 Branch `test/mdnotes-inbox-end-to-end`. Prove the whole path against the real vault
  rather than a `t.TempDir()`: with `MDNOTES_INBOX="$HOME/Documents/second-brain/00 - Inbox"`,
  run `org2md --yes` on a throwaway capture and confirm the note lands in `00 - Inbox/` and
  that no `Inbox/` directory appears at the vault root — the exact reproduction in issue #13.
  Then confirm the error path with a deliberately misspelled `MDNOTES_INBOX`, checking that the
  run exits non-zero, writes no Markdown beside the input, and creates no directory. Add
  whatever of this is automatable to `cmd/org2md/e2e_test.go` and verify with `make check`,
  `make test-race`, and `make cover`.

## Workflow follow-up

- Tag v2.0.0 once every task above has merged, since `$MDNOTES_DIR` support is removed.
- Close issue #13 referencing the merged pull requests.
- Archive the change, which syncs both delta specs into the main specs.
- Verify the archived result with `openspec validate --specs --strict` and
  `openspec list --specs`.
