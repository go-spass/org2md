---
title: "org2md: shipping the converter"
date: 2026-10-28
author: Chris
tags:
  - project
  - go
category: projects
id: 2026-10-28-org2md
---

The converter turns org files into Obsidian notes. See [[second-brain|the vault note]]
for how the inbox is organised.

# Parse the easy constructs #done #parser

Headings, paragraphs and *inline markup* all work. The tricky part was org's
word-boundary rule: without it `some_variable_name` loses a chunk.

## Constructs covered

- headings and paragraphs
- lists, including
  - nesting
  - [ ] checkboxes
  - [x] and finished ones
- **tables**: with alignment
- blocks

# Decide on the inbox layout #todo #priority/a #vault

**Scheduled:** 2026-10-29 09:00 +1w
**Deadline:** 2026-11-01

Options, roughly:

| Layout      | Effort | Note        |
| :---------- | :----: | ----------: |
| Flat Inbox  | low    | the default |
| Per-project | high   | tidier      |

```sh
# the move is a rename when it can be
mv note.md "$MDNOTES_DIR/Inbox/"
```

> Capture first, organise later.[^1]

---

# Write the README #next #docs

Nothing **fancy** needed, just the flags and the limits.

[^1]: The usual PARA advice.
