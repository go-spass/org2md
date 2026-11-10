# task-metadata Specification

## Purpose

Carries org TODO state, priority, and scheduling into Markdown while keeping the heading
structure that a list item could not hold. Derived from FR-6 in `docs/design.md`.

## Requirements

### Requirement: TODO keywords become status tags
A heading's TODO keyword SHALL be stripped from the title and appended to the heading line
as a status tag such as `#todo` or `#done`. The keyword set SHALL come from the file's
`#+TODO:`, `#+SEQ_TODO:`, or `#+TYP_TODO:` line when it has one, and otherwise from a default
list of the common workflow words.

#### Scenario: Default keyword set
- **WHEN** the source holds `* TODO Write the parser` and no TODO declaration
- **THEN** the output holds `# Write the parser #todo`

#### Scenario: Declared keyword set
- **WHEN** the file declares its own keywords via `#+TODO:` and a heading uses one of them
- **THEN** that keyword is recognized and rendered as its status tag

### Requirement: Heading tag order is fixed
A heading line SHALL carry its parts in the order title, status tag, priority tag, then the
author's own tags.

#### Scenario: Heading with every part
- **WHEN** a heading has a TODO keyword, an `[#A]` priority, and the tags `:work:idea:`
- **THEN** the output line is the title, then the status tag, then `#priority/a`, then
  `#work #idea`

### Requirement: Priority becomes a tag
An org priority cookie SHALL render as a `#priority/<letter>` tag, with the letter
lowercased.

#### Scenario: Priority A
- **WHEN** a heading carries `[#A]`
- **THEN** its line carries `#priority/a`

### Requirement: Scheduling lines open the body
`SCHEDULED:` and `DEADLINE:` lines under a heading SHALL render as `**Scheduled:** <date>`
and `**Deadline:** <date>` as the first lines of that heading's body. `CLOSED:` SHALL be
dropped.

#### Scenario: Scheduled heading
- **WHEN** a heading is followed by a `SCHEDULED: <2026-10-28 Wed>` line and a paragraph
- **THEN** the heading's body opens with `**Scheduled:** 2026-10-28`, followed by the
  paragraph

#### Scenario: Closed line dropped
- **WHEN** a heading carries a `CLOSED:` timestamp
- **THEN** it does not appear in the output

### Requirement: Time and repeater survive
A time of day or repeater in a scheduling timestamp SHALL be preserved after the date, and
only the redundant day name SHALL be dropped.

#### Scenario: Timestamp with time and repeater
- **WHEN** a heading carries `SCHEDULED: <2026-10-28 Wed 14:30 +1w>`
- **THEN** the body opens with `**Scheduled:** 2026-10-28 14:30 +1w`

### Requirement: TODO headings stay headings
A TODO heading SHALL remain a Markdown heading rather than collapsing into a `- [ ]` task
item, because it can own body content and subheadings that a list item cannot hold.

#### Scenario: TODO heading with children
- **WHEN** a TODO heading owns a paragraph and a subheading
- **THEN** the output keeps it as a heading with that paragraph and subheading beneath it
