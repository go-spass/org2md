# org2md — Functional Requirements & Initial Design

Status: draft, pre-implementation. Settled decisions are marked **Decided**; items still
open are collected in [Open questions](#open-questions).

## 1. Purpose

`org2md` converts a single Emacs org-mode file into a Markdown file suited to an Obsidian
vault. When `MDNOTES_DIR` is set, it offers to move the result into `$MDNOTES_DIR/Inbox`.

## 2. Settled decisions

| Area | Decision |
| --- | --- |
| Module path | `github.com/go-spass/org2md` |
| Layout | `cmd/org2md/` for CLI wiring, `internal/` for logic |
| CLI library | `spf13/cobra` — POSIX flags, flags after args, completions, room for subcommands |
| v1 org scope | Core syntax + YAML frontmatter + task metadata |
| Output | explicit `-o` wins; otherwise write beside the source and offer the move |
| Dates | plain text, no `[[date]]` daily-note links — the target vault has no daily notes |
| Underline | `<u>` — lossless, and the construct is rare enough that the HTML stays unobtrusive |
| Process | TDD; feature branches; `make` for build and test |

## 3. Functional requirements

### FR-1 — Invocation

```
org2md [flags] <file.org>
```

- Exactly one input file in v1. Zero or multiple arguments is a usage error.
- `-o, --output <path>` — write here instead of the default location. `-o -` writes to stdout.
- `--force` — overwrite an existing output file (without it, an existing file is an error).
- `-y, --yes` — assume yes to the move prompt; makes the tool safe for scripts and CI.
- `--no-move` — never move, even when `MDNOTES_DIR` is set.
- `-v, --verbose` — report conversion warnings and the final path on stderr.
- `--version` — version, commit, and build date.

### FR-2 — Output location

Resolution order:

1. `-o <path>` given → write exactly there. No prompt, no move.
2. No `-o` → write `<input-basename>.md` beside the input file, then:
   - `MDNOTES_DIR` set and `--no-move` absent → offer to move it to `$MDNOTES_DIR/Inbox/`
     (`-y` accepts without asking; a non-TTY stdin declines without asking).
   - otherwise → leave it beside the input.
3. Moving into a path that already exists is refused unless `--force`.

The prompt is interactive, so it lives in `main`/`internal/vault` and never in the
conversion path.

### FR-3 — Frontmatter

File-level org metadata becomes YAML frontmatter at the top of the Markdown:

| Org | Frontmatter |
| --- | --- |
| `#+TITLE:` | `title:` |
| `#+DATE:` | `date:` (normalized to `YYYY-MM-DD` when parseable, else verbatim) |
| `#+FILETAGS: :a:b:` | `tags: [a, b]` |
| `#+AUTHOR:` | `author:` |
| file-level `:PROPERTIES:` drawer | one key per property, key lowercased |

Recognized keywords are a whitelist; unrecognized `#+KEYWORD:` lines are dropped with a
warning. Frontmatter is omitted entirely when nothing maps to it.

### FR-4 — Structural conversion

| Org | Markdown |
| --- | --- |
| `*`, `**`, `***` headings | `#`, `##`, `###` (depth preserved) |
| Heading tags `:work:idea:` | `#work #idea` appended to the heading line |
| `-`, `+` lists | `-`, nesting and indentation preserved |
| `1.`, `1)` lists | `1.` |
| `- [ ]` / `- [X]` / `- [-]` | `- [ ]` / `- [x]` / `- [/]` |
| `- term :: definition` | `**term**: definition` |
| `#+BEGIN_SRC go` … `#+END_SRC` | fenced block with `go` info string |
| `#+BEGIN_EXAMPLE` … | fenced block, no language |
| `#+BEGIN_QUOTE` … | `>` blockquote |
| org table + `|---|` rule | GFM table with alignment row |
| `[fn:1]` / footnote definition | `[^1]` / `[^1]:` |
| `:LOGBOOK:` drawer | dropped |
| other drawers | dropped with a warning |
| `# comment` lines | dropped |
| horizontal rule `-----` | `---` |

### FR-5 — Inline conversion

| Org | Markdown |
| --- | --- |
| `*bold*` | `**bold**` |
| `/italic/` | `*italic*` |
| `+strike+` | `~~strike~~` |
| `=code=`, `~code~` | `` `code` `` |
| `_underline_` | `<u>underline</u>` (Obsidian renders inline HTML) |
| `[[https://x][label]]` | `[label](https://x)` |
| `[[https://x]]` | `<https://x>` |
| `[[file:note.org][label]]` | `[[note\|label]]` (Obsidian wikilink, `.org` stripped) |
| `[[*Some Heading]]` | `[[#Some Heading]]` |
| `<2026-10-28 Wed>` | `2026-10-28` |

Inline markup is **not** applied inside source, example, or verbatim content — a bare `*`
in a shell snippet must survive untouched. This is the main reason the design parses rather
than string-substitutes.

### FR-6 — Task metadata

- `* TODO Write the parser` → `## Write the parser` with the keyword stripped and a status
  tag (`#todo`, `#done`, `#next`, …) appended to the heading line.
- Priority `[#A]` → `#priority/a`.
- `SCHEDULED:`/`DEADLINE:` lines under a heading → `**Scheduled:** 2026-10-28` /
  `**Deadline:** 2026-10-28` as the first line of the heading's body.
- A time of day or repeater in the timestamp is preserved after the date:
  `<2026-10-28 Wed 14:30 +1w>` → `**Scheduled:** 2026-10-28 14:30 +1w`. The renderer splits
  the date from the remainder and drops only the redundant day name.
- `CLOSED:` → dropped.

Heading structure is preserved rather than collapsed into `- [ ]` list items: an org TODO
heading can own body content and subheadings, which a list item cannot hold without loss.
Org checkbox *list items* already match Markdown task syntax and convert directly (FR-4).

### FR-7 — Error handling

- Missing or unreadable input, bad argument count, or an existing output without `--force`
  → non-zero exit, message on stderr, no file written.
- Unrecognized org syntax never fails the run: the text passes through verbatim and a
  warning goes to stderr (shown only under `-v`, counted always).
- A non-`.org` extension is a warning, not an error.

## 4. Design

### Pipeline

```
file/stdin ──▶ org.Parse ──▶ *org.Document ──▶ md.Render ──▶ string
                                                               │
                                              vault.Place ◀────┘
                                          (write, prompt, move)
```

Conversion is pure: `io.Reader` in, string out, no filesystem and no environment access.
Everything interactive or environment-dependent sits in `internal/vault` and `main`, which
keeps the whole of FR-3 through FR-6 unit-testable without touching disk.

### Packages

| Package | Responsibility |
| --- | --- |
| `cmd/org2md` | Cobra root command, flag definitions, exit codes, wiring |
| `internal/org` | org source → AST; owns all org syntax knowledge |
| `internal/md` | AST → Markdown + frontmatter; owns all Obsidian conventions |
| `internal/vault` | `MDNOTES_DIR` resolution, output path choice, prompt, move |

Splitting `org` from `md` keeps the two dialects independent — an org construct with no
Markdown equivalent is a rendering decision, not a parsing one, and a second output format
later would not touch the parser.

### Parsing approach

Line-oriented, two phases:

1. **Block phase** — classify each line (heading, keyword, block delimiter, list item,
   table row, drawer, blank) and build the block-level AST. Org is overwhelmingly
   line-based, so this needs no grammar; block delimiters like `#+BEGIN_SRC` simply switch
   the scanner into a raw-capture mode, which is what gives FR-5 its no-transform
   guarantee.
2. **Inline phase** — run emphasis/link/timestamp recognition over the text of nodes that
   permit markup. Code and verbatim nodes are skipped.

### AST sketch

```go
type Document struct {
    Meta       map[string]string // #+KEYWORD
    Properties map[string]string // file-level :PROPERTIES:
    Children   []Node
}

// Blocks: Heading, Paragraph, List, ListItem, CodeBlock, Quote, Example,
//         Table, FootnoteDef, Rule
// Inlines: Text, Bold, Italic, Underline, Strike, Code, Link,
//          Timestamp, FootnoteRef, Tag
```

`Heading` carries `Level`, `Title []Inline`, `Todo`, `Priority`, `Tags`, `Scheduled`,
`Deadline`, and `Children`, so FR-6 is a rendering concern over data the parser already has.

### Testing strategy

- Table-driven unit tests per construct in `internal/org` and `internal/md` — the common
  case is a short org snippet and its expected Markdown.
- Golden files in `testdata/` for whole documents, refreshable with a `-update` flag.
- `internal/vault` tested against `t.TempDir()` with `MDNOTES_DIR` set via `t.Setenv`; the
  prompt reads from an injected `io.Reader` so the move logic needs no TTY.
- One end-to-end test that runs the built binary over a realistic org note.

### Makefile

`build`, `test`, `test-race`, `cover`, `vet`, `fmt`, `clean`, plus `install`. Version
information is stamped via `-ldflags -X main.version=…`.

## 5. Implementation plan

Each step is red-green-refactor, gets its own branch and PR, and leaves the tree building
and tests passing.

| # | Step | Done when |
| --- | --- | --- |
| 0 | ✅ Scaffold: `go.mod`, `Makefile`, Cobra root command, `--version` | `make build && make test` pass |
| 1 | ✅ `internal/org`: headings + paragraphs | parser produces the AST for a nested-heading document |
| 2 | ✅ `internal/md`: headings + paragraphs, golden harness | first end-to-end snippet converts |
| 3 | ✅ Inline phase: emphasis, code, links, timestamps (FR-5) | word-boundary rules hold; inline verbatim keeps its markers |
| 4 | ✅ Lists, checkboxes, definition lists | nested mixed lists round-trip |
| 5 | Blocks: src, example, quote, rules | language info string preserved; emphasis inside a src block provably untouched |
| 6 | Tables and footnotes | GFM alignment row correct |
| 7 | Frontmatter (FR-3) | title/date/tags/properties emitted, empty case omitted |
| 8 | Task metadata (FR-6) | TODO, priority, scheduled/deadline rendered |
| 9 | `internal/vault` + CLI flags (FR-1, FR-2) | prompt, `-y`, `--no-move`, `--force`, `-o -` |
| 10 | End-to-end golden test on a real note; README refresh | realistic org file converts cleanly |

Steps 1–8 need no CLI and no filesystem, so the bulk of the tool is testable from step 1.

## Open questions

1. **Status tags vs. callouts for TODO headings** — `#todo` is simple and searchable;
   Obsidian callouts or the Tasks plugin's emoji format are richer but tie output to
   specific plugins. Current choice: plain tags.
2. **`:LOGBOOK:` and clock entries** — dropped in v1; worth a `--keep-drawers` flag later?
3. **Multiple input files / directory mode** — out of scope for v1, but it is the obvious
   first subcommand and a reason Cobra was chosen.
4. **Continuation text interleaved with a nested list** — `ListItem` keeps its own text in
   `Lines` and its nested lists in `Children`, so an item holding both renders all of its
   text before the nested list even when the source interleaved them. Fixing it means
   modelling item content as one ordered `[]Node`, which also changes whether such an item
   renders tight or loose. Rare enough to defer, common enough to write down.

Resolved: scheduled/deadline dates render as plain text (no daily-note links — the vault
does not use daily notes, so linking would only mint unresolved links), and `_underline_`
renders as `<u>`.
