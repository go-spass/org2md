# warning-reporting Specification

## Purpose

Governs what org2md treats as fatal, what it reports as a warning, and what it never drops
in silence. Derived from FR-7 in `docs/design.md`.

## Requirements

### Requirement: Fatal conditions write nothing
Missing or unreadable input, a bad argument count, an existing output without `--force`, or a
configured `$MDNOTES_INBOX` that is not an existing directory SHALL exit non-zero with a
message on stderr and SHALL leave no file written.

#### Scenario: Unreadable input
- **WHEN** `org2md missing.org` runs and the file does not exist
- **THEN** a message goes to stderr, no file is written, and the exit status is non-zero

#### Scenario: Unusable inbox
- **WHEN** `MDNOTES_INBOX` names a path that is not an existing directory and
  `org2md capture.org` runs
- **THEN** a message naming that path goes to stderr, no file is written, and the exit status
  is non-zero

### Requirement: Unrecognized syntax never fails the run
Org syntax org2md does not recognize SHALL NOT fail the conversion: the run succeeds, the
construct raises a warning, and body text org2md cannot map SHALL reach the output verbatim
rather than being discarded.

#### Scenario: Unknown block-level text
- **WHEN** the source holds a construct org2md has no mapping for
- **THEN** the run succeeds, the text appears in the output unchanged, and a warning names
  the construct

#### Scenario: Keyword with no frontmatter equivalent
- **WHEN** the source holds an unrecognized `#+SOMETHING:` line
- **THEN** the run succeeds, the line is dropped rather than rendered as body text, and a
  warning names the dropped keyword

### Requirement: Warnings are counted always and detailed under verbose
Warnings SHALL be collected for every run. Under `-v, --verbose` each SHALL be printed on
stderr in document order; without it, a single count line SHALL be printed on stderr naming
how many constructs could not be converted and pointing at `--verbose`.

#### Scenario: Verbose run
- **WHEN** a conversion producing two warnings runs with `--verbose`
- **THEN** both warnings and the path written are reported on stderr, in document order

#### Scenario: Quiet run
- **WHEN** the same conversion runs without `--verbose`
- **THEN** stderr carries one line naming the warning count and `--verbose`, no individual
  warning, and the exit status reflects success

#### Scenario: Clean run
- **WHEN** a conversion produces no warnings and runs without `--verbose`
- **THEN** nothing is written to stderr

### Requirement: Anything Markdown cannot say is reported
Conversion SHALL return its warnings alongside the Markdown, so every construct org can
express and Markdown cannot is reported rather than dropped silently. The `:LOGBOOK:` drawer
is the one deliberate exception.

#### Scenario: Non-logbook drawer
- **WHEN** a heading owns a `:NOTES:` drawer
- **THEN** the drawer is dropped and a warning names it

#### Scenario: Logbook drawer
- **WHEN** a heading owns a `:LOGBOOK:` drawer
- **THEN** the drawer is dropped and no warning is raised

### Requirement: A wrong extension is only a warning
An input file whose extension is not `.org` SHALL be converted, with a warning, rather than
refused.

#### Scenario: Text file input
- **WHEN** `org2md notes.txt` runs on a file holding org syntax
- **THEN** the conversion proceeds and a warning names the unexpected extension
