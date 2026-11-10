# Spec Delta

## MODIFIED Requirements

### Requirement: Explicit output wins
When `-o <path>` is given, org2md SHALL write exactly there, and SHALL neither prompt nor
offer to move the result. It SHALL NOT read or validate `$MDNOTES_INBOX` on such a run.

#### Scenario: Explicit path
- **WHEN** `org2md -o ~/notes/note.md capture.org` runs with `MDNOTES_INBOX` set
- **THEN** the Markdown is written to `~/notes/note.md` and no move prompt appears

#### Scenario: Explicit path with an unusable inbox
- **WHEN** `org2md -o ~/notes/note.md capture.org` runs with `MDNOTES_INBOX` naming a
  directory that does not exist
- **THEN** the Markdown is written to `~/notes/note.md` and the run succeeds, because the
  inbox is irrelevant to it

### Requirement: Offer the vault move
After writing beside the input, org2md SHALL offer to move the file into `$MDNOTES_INBOX/`
when `MDNOTES_INBOX` is set and `--no-move` is absent. The variable names the destination
directory itself, and org2md SHALL NOT append a subdirectory to it. `-y` SHALL accept without
asking, and a non-interactive stdin SHALL decline without asking. Otherwise the file SHALL
stay beside the input.

#### Scenario: Prompt accepted
- **WHEN** `MDNOTES_INBOX` is set and the user answers yes to the prompt
- **THEN** the file is moved to `$MDNOTES_INBOX/<basename>.md` and that path is the one
  reported

#### Scenario: Destination is the variable itself
- **WHEN** `MDNOTES_INBOX` is `/vault/00 - Inbox` and the move is accepted
- **THEN** the file lands at `/vault/00 - Inbox/<basename>.md`, and no `Inbox` subdirectory is
  appended or created

#### Scenario: Move suppressed
- **WHEN** `org2md --no-move capture.org` runs with `MDNOTES_INBOX` set
- **THEN** no prompt appears and the file stays beside the input

#### Scenario: Non-interactive stdin
- **WHEN** `MDNOTES_INBOX` is set, `-y` is absent, and stdin is not a terminal
- **THEN** the move is declined without prompting and the file stays beside the input

#### Scenario: No inbox configured
- **WHEN** `MDNOTES_INBOX` is unset or empty
- **THEN** no prompt appears and the file stays beside the input

## ADDED Requirements

### Requirement: An unusable inbox is an error
When a move is in play, org2md SHALL require `$MDNOTES_INBOX` to name an existing directory.
A path that does not exist, or that exists but is not a directory, SHALL exit non-zero with a
message naming the path. org2md SHALL NOT create the directory.

#### Scenario: Inbox does not exist
- **WHEN** `MDNOTES_INBOX` names a path that does not exist and `org2md capture.org` runs
- **THEN** the run exits non-zero, stderr names the configured path, and no directory is
  created anywhere

#### Scenario: Inbox is a regular file
- **WHEN** `MDNOTES_INBOX` names an existing regular file
- **THEN** the run exits non-zero and stderr names the configured path

#### Scenario: Typo creates nothing
- **WHEN** `MDNOTES_INBOX` is `/vault/00 - Inbx` and the real directory is `/vault/00 - Inbox`
- **THEN** the run errors, and `/vault/00 - Inbx` does not exist afterwards

### Requirement: The inbox is validated before anything is written
org2md SHALL check the configured inbox before writing the Markdown beside the input, so an
unusable inbox leaves no file behind. It SHALL NOT ask the move question when the destination
cannot be honored.

#### Scenario: Nothing is left beside the input
- **WHEN** `MDNOTES_INBOX` names a missing directory and `org2md /tmp/capture.org` runs
- **THEN** the run errors and `/tmp/capture.md` does not exist

#### Scenario: No prompt for an unusable destination
- **WHEN** the same run has an interactive stdin and `-y` is absent
- **THEN** no move question is asked before the error

### Requirement: MDNOTES_DIR is not consulted
org2md SHALL NOT read, validate, or report on `$MDNOTES_DIR`. The variable belongs to other
tools, and its presence or absence SHALL have no effect on where org2md writes.

#### Scenario: Only the old variable is set
- **WHEN** `MDNOTES_DIR` is set, `MDNOTES_INBOX` is unset, and `org2md capture.org` runs
- **THEN** no prompt appears, the file stays beside the input, and the run succeeds with no
  message about `MDNOTES_DIR`

#### Scenario: Old variable names a missing directory
- **WHEN** `MDNOTES_DIR` names a path that does not exist and `MDNOTES_INBOX` names a real one
- **THEN** the move is offered against `$MDNOTES_INBOX` and `$MDNOTES_DIR` is never checked
