# cli-invocation Specification

## Purpose

Defines how org2md is called from a shell: the argument contract, the flags that steer a
conversion, and the exit status each outcome carries. Derived from FR-1 in `docs/design.md`.

## Requirements

### Requirement: Single input file
org2md SHALL accept at most one positional argument, the org file to convert. More than one
argument SHALL be an argument error naming the count received and pointing at `--help`.

#### Scenario: One org file
- **WHEN** `org2md capture.org` runs
- **THEN** that file is converted and the command exits 0

#### Scenario: Two or more files
- **WHEN** `org2md a.org b.org` runs
- **THEN** no file is written, a message naming the argument count and `org2md --help` goes
  to stderr, and the exit status is non-zero

### Requirement: A bare run explains itself
Run with no arguments, org2md SHALL print its help — usage, examples, and flags — on stdout
and exit 0. A bare invocation is a question, not a usage error.

#### Scenario: No arguments
- **WHEN** `org2md` runs with no arguments and no flags
- **THEN** usage, the examples, and the flag list are written to stdout and the exit status
  is 0

### Requirement: Conversion flags
org2md SHALL offer `-o, --output <path>` to choose the destination, `--force` to overwrite
an existing file, `-y, --yes` to accept the move prompt without asking, `--no-move` to never
move, and `-v, --verbose` to report warnings and the final path on stderr.

#### Scenario: Output to stdout
- **WHEN** `org2md -o - capture.org` runs
- **THEN** the Markdown is written to stdout, no file is created, and no move is offered

#### Scenario: Quiet by default
- **WHEN** a conversion that produces warnings runs without `-v`
- **THEN** the individual warnings are not printed, only a count line pointing at
  `--verbose`, and the exit status reflects success

### Requirement: Version reporting
`--version` SHALL report the version, commit, and build date. When the linker did not stamp
a version, org2md SHALL fall back to Go's embedded build information rather than reporting
the `dev` placeholder.

#### Scenario: Stamped build
- **WHEN** `org2md --version` runs on a build whose version was stamped via ldflags
- **THEN** the stamped version, commit, and build date are printed

#### Scenario: go install build
- **WHEN** `org2md --version` runs on a binary produced by `go install`
- **THEN** the module version and VCS revision from the embedded build info are printed
  instead of `dev`
