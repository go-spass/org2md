# org2md
Go CLI that converts Emacs org-mode files to Markdown

## Overview
`org2md` is a CLI tool written in Go that will take the specified org-mode
file and generate a Markdown file capable of being added to Obsidian. If the
environment variable `MDNOTES_DIR` is defined, the tool will offer to move the
generated Markdown file into "$MDNOTES_DIR/Inbox".

