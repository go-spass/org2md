# org2md

Go CLI that converts Emacs org-mode files to Markdown for an Obsidian vault.

`org2md` reads an org file and writes a Markdown file that an Obsidian vault can use
directly: org metadata becomes YAML frontmatter, `TODO` keywords and org tags become
Obsidian tags, and links between org files become wikilinks. When `MDNOTES_DIR` is set, it
offers to file the result in `$MDNOTES_DIR/Inbox`.

## Install

```sh
go install github.com/go-spass/org2md/cmd/org2md@latest
```

Or build from a clone, which puts the binary in `./bin`:

```sh
make build
```

Either way `--version` reports a real version: `make build` stamps it with `-ldflags`, and a
`go install` build reads the module version and commit from the build information Go embeds
in every binary.

## Usage

```sh
org2md [flags] <file.org>
```

With no `--output`, the Markdown is written beside the input file — `notes.org` becomes
`notes.md`. If `MDNOTES_DIR` is set, `org2md` then asks whether to move it into
`$MDNOTES_DIR/Inbox`.

Run `org2md` with no arguments (or `org2md --help`) to print this usage, the flags below and
worked examples.

| Flag | Meaning |
| --- | --- |
| `-o, --output <path>` | Write here instead of beside the input. `-o -` writes to stdout and never moves the file. |
| `--force` | Overwrite an existing file. Without it, an existing destination is an error and nothing is written. |
| `-y, --yes` | Answer yes to the move prompt. |
| `--no-move` | Never move into `$MDNOTES_DIR/Inbox`. |
| `-v, --verbose` | Report each conversion warning and the path written. |
| `--version` | Version, commit and build date. |

The prompt is only asked when stdin is a terminal, so a piped or redirected run declines the
move rather than hanging. The prompt itself goes to stderr, which keeps `-o -` pipeable.

## What it converts

Headings, paragraphs, and the emphasis markers `*bold*`, `/italic/`, `_underline_`,
`+strike+`, `=verbatim=` and `~code~`. Org's word-boundary rule is respected, so
`some_variable_name`, `and/or` and `2*3*4` come through untouched.

Lists, with nesting, ordered and unordered markers, definition items (`- term :: text`) and
checkboxes — `[-]` becomes `[/]`, which Obsidian understands as partially done.

Source, example and quote blocks. Source-block content is never reinterpreted, so a bare `*`
or a line starting with `-` inside a shell snippet survives. The fence grows longer than any
run of backticks in the content.

Tables, including org's alignment cookies (`| <l> | <r> | <c> |`), which become a GFM
alignment row. Footnotes, horizontal rules, and timestamps.

Links: `[[url][label]]` becomes a Markdown link, a bare `[[url]]` becomes `<url>`, and a
link to another org file becomes a wikilink to the converted note of that name.

### Frontmatter

`#+TITLE`, `#+DATE`, `#+AUTHOR` and `#+FILETAGS` become `title`, `date`, `author` and
`tags`. A `:PROPERTIES:` drawer before the first heading contributes one key each, lowercased.
Values are quoted only when they need it, and a date is reduced to the ISO date inside it, so
`#+DATE: <2026-10-28 Wed>` becomes `date: 2026-10-28`.

### Tasks

A headline's keyword, priority and tags become tags on the Markdown heading, so
`* TODO [#A] Fix the parser :work:bug:` becomes:

```
# Fix the parser #todo #priority/a #work #bug
```

The keyword set comes from the file's `#+TODO:` line when it has one. `SCHEDULED:` and
`DEADLINE:` become the first lines of the heading's body, keeping any time or repeater.

## Warnings

Anything org can express that Markdown cannot is reported rather than dropped in silence.
The count always appears; `--verbose` lists the detail:

```
org2md: 2 constructs could not be converted; run with --verbose for details
```

A `:LOGBOOK:` drawer and a `CLOSED:` date are the deliberate quiet exceptions: they are
bookkeeping with no equivalent. Org syntax `org2md` does not recognize is passed through
verbatim rather than mangled, and never fails a conversion.

## Limitations

- One input file per run. A directory mode would be the first subcommand.
- Only a `:PROPERTIES:` drawer before the first heading becomes frontmatter; a per-heading
  one has nowhere to go in a single YAML block and is reported as dropped.
- A list item holding both its own continuation text and a nested list renders the text
  first, whatever the source order.
- Inactive timestamps (`[2026-10-28 Wed]`) and org's inline footnote form (`[fn:1:text]`)
  pass through as text.

The open questions at the end of [docs/design.md](docs/design.md) track these.

## Development

```sh
make build        # build ./bin/org2md
make test         # go test ./...
make test-race
make cover
make check        # vet + test
make golden       # rewrite the golden files in testdata
make fmt
```

[docs/design.md](docs/design.md) holds the functional requirements, the org to Markdown
mapping tables, and the package design. [CLAUDE.md](CLAUDE.md) records the project's
conventions: test-driven, one branch and PR per plan step, `make` for everything.
