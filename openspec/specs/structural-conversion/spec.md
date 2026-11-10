# structural-conversion Specification

## Purpose

Converts org block-level structure — headings, lists, blocks, tables, footnotes, rules — into
the Markdown an Obsidian vault renders. Derived from FR-4 in `docs/design.md`.

## Requirements

### Requirement: Headings preserve depth
Org heading stars SHALL become ATX headings of the same depth, and a heading's own tags
SHALL be appended to its line as hashtags.

#### Scenario: Nested headings with tags
- **WHEN** the source holds `* Top` and `** Child :work:idea:`
- **THEN** the output holds `# Top` and `## Child #work #idea`

### Requirement: Lists keep their kind and nesting
`-` and `+` bullets SHALL render as `-`, `1.` and `1)` as `1.`, and nesting and indentation
SHALL be preserved. Definition items `- term :: definition` SHALL render as
`**term**: definition`.

#### Scenario: Mixed nested list
- **WHEN** an ordered list contains a nested bullet list
- **THEN** both survive with their markers and indentation intact

#### Scenario: Definition item
- **WHEN** the source holds `- parser :: reads org`
- **THEN** the output holds `**parser**: reads org`

### Requirement: Checkbox items map to task syntax
Org checkbox items SHALL render as Markdown task items: `- [ ]` unchanged, `- [X]` as
`- [x]`, and `- [-]` as `- [/]`.

#### Scenario: All three states
- **WHEN** a list holds an unchecked, a checked, and a partial item
- **THEN** the output holds `- [ ]`, `- [x]`, and `- [/]` respectively

### Requirement: Blocks become fences and quotes
`#+BEGIN_SRC <lang>` SHALL become a fenced code block carrying `<lang>` as its info string,
`#+BEGIN_EXAMPLE` a fence with no language, and `#+BEGIN_QUOTE` a `>` blockquote. A
horizontal rule `-----` SHALL become `---`.

#### Scenario: Source block keeps its language
- **WHEN** the source holds a `#+BEGIN_SRC go` block
- **THEN** the output holds a fence opened with `` ```go ``

#### Scenario: Example block has no language
- **WHEN** the source holds a `#+BEGIN_EXAMPLE` block
- **THEN** the output holds a bare fence

### Requirement: Tables render as GFM
An org table SHALL render as a GFM table. A `|---|` rule SHALL produce the alignment row,
and an org alignment cookie such as `| <l> |` SHALL set that column's alignment.

#### Scenario: Table with a rule
- **WHEN** the source holds a header row, a `|---|` rule, and a body row
- **THEN** the output holds the header, a correct GFM alignment row, and the body row

#### Scenario: Left-alignment cookie
- **WHEN** a column carries the `<l>` cookie
- **THEN** that column's alignment cell is `:---`

### Requirement: Footnotes map to GFM footnotes
Org footnote references SHALL render as `[^n]` and footnote definitions as `[^n]:`.

#### Scenario: Reference and definition
- **WHEN** the source holds `[fn:1]` in a paragraph and its definition below
- **THEN** the output holds `[^1]` and a `[^1]:` definition line

### Requirement: Comments and logbooks are dropped
`# comment` lines and `:LOGBOOK:` drawers SHALL be dropped from the output. The LOGBOOK drop
SHALL be silent; any other drawer SHALL be dropped with a warning.

#### Scenario: Comment line
- **WHEN** a paragraph is preceded by a `# note to self` line
- **THEN** that line does not appear in the output and no warning is reported

#### Scenario: Logbook drawer
- **WHEN** a heading owns a `:LOGBOOK:` drawer with clock entries
- **THEN** the drawer is absent from the output and no warning is reported
