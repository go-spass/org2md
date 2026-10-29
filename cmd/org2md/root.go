package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

type buildInfo struct {
	version string
	commit  string
	date    string
}

var errNotImplemented = errors.New("conversion is not implemented yet")

func newRootCmd(b buildInfo) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org2md [flags] <file.org>",
		Short: "Convert an Emacs org-mode file to Obsidian-ready Markdown",
		Long: `Convert an Emacs org-mode file to a Markdown file suited to an Obsidian vault.

With no --output, the Markdown is written beside the input file. When MDNOTES_DIR is
set, org2md then offers to move the result into $MDNOTES_DIR/Inbox.`,
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errNotImplemented
		},
	}

	cmd.Version = b.version
	cmd.SetVersionTemplate(fmt.Sprintf("org2md %s (commit %s, built %s)\n", b.version, b.commit, b.date))

	return cmd
}
