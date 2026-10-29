package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/go-spass/org2md/internal/md"
	"github.com/go-spass/org2md/internal/org"
	"github.com/go-spass/org2md/internal/vault"
)

// notesDirEnv names the vault org2md offers to file converted notes into.
const notesDirEnv = "MDNOTES_DIR"

type buildInfo struct {
	version string
	commit  string
	date    string
}

type options struct {
	output  string
	force   bool
	yes     bool
	noMove  bool
	verbose bool
}

func newRootCmd(b buildInfo) *cobra.Command {
	var o options

	cmd := &cobra.Command{
		Use:   "org2md [flags] <file.org>",
		Short: "Convert an Emacs org-mode file to Obsidian-ready Markdown",
		Long: fmt.Sprintf(`Convert an Emacs org-mode file to a Markdown file suited to an Obsidian vault.

With no --output, the Markdown is written beside the input file. When %s is
set, org2md then offers to move the result into $%s/Inbox.`, notesDirEnv, notesDirEnv),
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return convert(cmd, args[0], &o)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&o.output, "output", "o", "", `write here instead of beside the input; "-" writes to stdout`)
	f.BoolVar(&o.force, "force", false, "overwrite an existing file")
	f.BoolVarP(&o.yes, "yes", "y", false, "answer yes to the move prompt")
	f.BoolVar(&o.noMove, "no-move", false, "never move into $"+notesDirEnv+"/Inbox")
	f.BoolVarP(&o.verbose, "verbose", "v", false, "report conversion warnings and the path written")

	cmd.Version = b.version
	cmd.SetVersionTemplate(fmt.Sprintf("org2md %s (commit %s, built %s)\n", b.version, b.commit, b.date))

	return cmd
}

func convert(cmd *cobra.Command, input string, o *options) error {
	stderr := cmd.ErrOrStderr()

	// A wrong extension is worth saying out loud, but the file may well still
	// be org, so it is not a refusal.
	if filepath.Ext(input) != ".org" {
		fmt.Fprintf(stderr, "org2md: %s does not look like an org file\n", input)
	}

	f, err := os.Open(input)
	if err != nil {
		return err
	}
	defer f.Close()

	doc, err := org.Parse(f)
	if err != nil {
		return err
	}

	content, warnings := md.Render(doc)
	report(stderr, warnings, o.verbose)

	// The prompt goes to stderr so that "-o -" stays pipeable.
	placer := &vault.Placer{
		NotesDir: os.Getenv(notesDirEnv),
		Force:    o.force,
		Yes:      o.yes,
		NoMove:   o.noMove,
		Stdout:   cmd.OutOrStdout(),
		Prompt:   stderr,
		Answers:  answerReader(cmd.InOrStdin()),
	}

	path, err := placer.Place(input, o.output, content)
	if err != nil {
		return err
	}
	if o.verbose && path != "" {
		fmt.Fprintf(stderr, "org2md: wrote %s\n", path)
	}

	return nil
}

// report prints conversion warnings. They are always counted, so nothing is
// lost quietly, but the detail waits for --verbose.
func report(w io.Writer, warnings []string, verbose bool) {
	if len(warnings) == 0 {
		return
	}

	if verbose {
		for _, warning := range warnings {
			fmt.Fprintf(w, "org2md: %s\n", warning)
		}
		return
	}

	noun := "constructs"
	if len(warnings) == 1 {
		noun = "construct"
	}
	fmt.Fprintf(w, "org2md: %d %s could not be converted; run with --verbose for details\n", len(warnings), noun)
}

// answerReader reports where a prompt may read its answer, or nil when there
// is no one to ask: a redirected or piped stdin declines the move rather than
// consuming a line of whatever was piped in.
//
// This asks term.IsTerminal rather than testing os.ModeCharDevice, because
// /dev/null is a character device too, and the usual idiom would print a
// prompt into a pipeline that can never answer it.
func answerReader(in io.Reader) io.Reader {
	if f, ok := in.(*os.File); ok && !term.IsTerminal(int(f.Fd())) {
		return nil
	}
	return in
}
