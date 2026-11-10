# inline-conversion Specification

## Purpose

Converts org inline markup — emphasis, code, links, timestamps — into Markdown, without ever
touching text inside verbatim content. Derived from FR-5 in `docs/design.md`.

## Requirements

### Requirement: Emphasis maps to Markdown
`*bold*` SHALL render as `**bold**`, `/italic/` as `*italic*`, `+strike+` as `~~strike~~`,
`=code=` and `~code~` as inline code, and `_underline_` as `<u>underline</u>`, which Obsidian
renders as inline HTML.

#### Scenario: Each emphasis form
- **WHEN** a paragraph holds one of each form
- **THEN** each is replaced by its Markdown or HTML equivalent

#### Scenario: Markers need word boundaries
- **WHEN** a paragraph holds `2*3*4` or an intra-word underscore
- **THEN** no emphasis is applied and the text passes through unchanged

### Requirement: Links map by target
A labelled URL link SHALL render as `[label](url)`, a bare URL link as `<url>`, a
`file:note.org` link as an Obsidian wikilink with the `.org` extension stripped, and an
in-document `[[*Some Heading]]` link as `[[#Some Heading]]`.

#### Scenario: Labelled and bare URLs
- **WHEN** the source holds `[[https://x][label]]` and `[[https://x]]`
- **THEN** the output holds `[label](https://x)` and `<https://x>`

#### Scenario: File link becomes a wikilink
- **WHEN** the source holds `[[file:note.org][label]]`
- **THEN** the output holds `[[note|label]]`

### Requirement: Timestamps render as plain dates
An inline org timestamp SHALL render as its date in `YYYY-MM-DD` form, as plain text. The
target vault does not use daily notes, so a date SHALL NOT become a link.

#### Scenario: Inline active timestamp
- **WHEN** a paragraph holds `<2026-10-28 Wed>`
- **THEN** the output holds `2026-10-28` as plain text, with no brackets and no link

### Requirement: Verbatim content survives untransformed
Inline markup SHALL NOT be applied inside source, example, or inline verbatim content. A
bare `*`, `/`, or `_` in a snippet SHALL reach the output byte for byte.

#### Scenario: Shell glob in a source block
- **WHEN** a `#+BEGIN_SRC sh` block holds `rm *.tmp`
- **THEN** the fenced output holds `rm *.tmp` with no emphasis applied

#### Scenario: Inline verbatim keeps its content
- **WHEN** a paragraph holds `=a*b*c=`
- **THEN** the inline code span holds `a*b*c` unchanged
