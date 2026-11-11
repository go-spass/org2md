# output-placement Specification

## Purpose

Decides where the converted Markdown lands: beside the input, at an explicit path, on
stdout, or moved into the configured inbox directory of an Obsidian vault. Derived from FR-2
in `docs/design.md`.

## Requirements

### Requirement: Explicit output wins
When `-o <path>` is given, org2md SHALL write exactly there, and SHALL neither prompt nor
offer to move the result.

#### Scenario: Explicit path
- **WHEN** `org2md -o ~/notes/note.md capture.org` runs with `MDNOTES_DIR` set
- **THEN** the Markdown is written to `~/notes/note.md` and no move prompt appears

### Requirement: Default location beside the input
With no `-o`, org2md SHALL write `<input-basename>.md` in the input file's directory.

#### Scenario: No output flag
- **WHEN** `org2md /tmp/capture.org` runs
- **THEN** the Markdown is written to `/tmp/capture.md`

### Requirement: Offer the vault move
After writing beside the input, org2md SHALL offer to move the file into
`$MDNOTES_DIR/Inbox/` when `MDNOTES_DIR` is set and `--no-move` is absent. `-y` SHALL accept
without asking, and a non-interactive stdin SHALL decline without asking. Otherwise the file
SHALL stay beside the input.

#### Scenario: Prompt accepted
- **WHEN** `MDNOTES_DIR` is set and the user answers yes to the prompt
- **THEN** the file is moved to `$MDNOTES_DIR/Inbox/<basename>.md` and that path is the one
  reported

#### Scenario: Move suppressed
- **WHEN** `org2md --no-move capture.org` runs with `MDNOTES_DIR` set
- **THEN** no prompt appears and the file stays beside the input

#### Scenario: Non-interactive stdin
- **WHEN** `MDNOTES_DIR` is set, `-y` is absent, and stdin is not a terminal
- **THEN** the move is declined without prompting and the file stays beside the input

### Requirement: Existing files are not overwritten
Writing or moving onto a path that already exists SHALL be refused unless `--force` is
given, and the refusal SHALL leave the existing file untouched.

#### Scenario: Output already exists
- **WHEN** `capture.md` already exists and `org2md capture.org` runs without `--force`
- **THEN** nothing is written, the existing file is unchanged, and the exit status is
  non-zero

#### Scenario: Forced overwrite
- **WHEN** the same run is given `--force`
- **THEN** the existing file is replaced with the converted Markdown
