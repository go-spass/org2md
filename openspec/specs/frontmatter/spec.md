# frontmatter Specification

## Purpose

Turns file-level org metadata into the YAML frontmatter block an Obsidian vault reads.
Derived from FR-3 in `docs/design.md`.

## Requirements

### Requirement: Recognized keywords become frontmatter
org2md SHALL map `#+TITLE:` to `title:`, `#+DATE:` to `date:`, `#+AUTHOR:` to `author:`, and
`#+FILETAGS: :a:b:` to `tags: [a, b]`, emitting them as a YAML block at the top of the
Markdown.

#### Scenario: Title, date, author, and filetags
- **WHEN** a document carries `#+TITLE: Notes`, `#+DATE: 2026-10-28`, `#+AUTHOR: Chris`, and
  `#+FILETAGS: :work:idea:`
- **THEN** the output opens with a YAML block holding `title`, `date`, `author`, and a `tags`
  list of `work` and `idea`

### Requirement: Dates are normalized when parseable
A `#+DATE:` value SHALL be normalized to `YYYY-MM-DD` when it can be parsed, and passed
through verbatim when it cannot.

#### Scenario: Unparseable date
- **WHEN** the document carries `#+DATE: sometime in autumn`
- **THEN** `date:` holds that text unchanged

### Requirement: File-level properties become keys
A file-level `:PROPERTIES:` drawer SHALL contribute one frontmatter key per property, with
the key lowercased.

#### Scenario: Property drawer before the first heading
- **WHEN** a file-level drawer declares `:Status: active`
- **THEN** the frontmatter carries `status: active`

### Requirement: Frontmatter is omitted when empty
org2md SHALL emit no frontmatter block at all when nothing in the document maps to it.

#### Scenario: Document with no metadata
- **WHEN** a document has only headings and paragraphs
- **THEN** the output starts with the first heading and contains no `---` delimiters
