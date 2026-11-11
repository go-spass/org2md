# Spec Delta

## MODIFIED Requirements

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
