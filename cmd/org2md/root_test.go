package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleOrg = "#+TITLE: Sample\n#+OPTIONS: toc:nil\n\n* TODO A task\nBody *text*.\n"

const sampleMarkdown = "---\ntitle: Sample\n---\n\n# A task #todo\n\nBody **text**.\n"

// run executes a fresh root command and captures both streams.
func run(t *testing.T, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	cmd := newRootCmd(buildInfo{version: "1.2.3", commit: "abc1234", date: "2026-10-28T00:00:00Z"})
	var out, errs bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errs)
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetArgs(args)

	err = cmd.Execute()
	return out.String(), errs.String(), err
}

// orgFile writes a sample org file in a fresh directory.
func orgFile(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "note.org")
	if err := os.WriteFile(path, []byte(sampleOrg), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return string(b)
}

func TestVersion(t *testing.T) {
	out, _, err := run(t, "", "--version")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	for _, want := range []string{"org2md", "1.2.3", "abc1234", "2026-10-28T00:00:00Z"} {
		if !strings.Contains(out, want) {
			t.Errorf("--version output is missing %q, got %q", want, out)
		}
	}
}

// -v is verbose, so Cobra leaves --version without a shorthand. Converting a
// file with -v must convert rather than print the version.
func TestShortVIsVerbose(t *testing.T) {
	input := orgFile(t)

	_, stderr, err := run(t, "", "-v", input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if strings.Contains(stderr, "1.2.3") {
		t.Errorf("-v printed the version, want verbose output: %q", stderr)
	}
	if !strings.Contains(stderr, "wrote") {
		t.Errorf("-v stderr = %q, want it to report the path written", stderr)
	}
}

// A bare org2md is someone asking what the tool takes, not an error: the help
// goes to stdout so it can be paged or grepped, and the run succeeds.
func TestNoArgumentsPrintsHelp(t *testing.T) {
	stdout, stderr, err := run(t, "")
	if err != nil {
		t.Fatalf("Execute() error = %v, want the help and no error", err)
	}

	wants := []string{
		"org2md [flags] <file.org>",
		"Examples:",
		"org2md capture.org",
		"--output",
		"--force",
		"--yes",
		"--no-move",
		"--verbose",
		"--version",
	}
	for _, want := range wants {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout is missing %q, got:\n%s", want, stdout)
		}
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want the help on stdout alone", stderr)
	}
}

func TestHelpShowsExamples(t *testing.T) {
	stdout, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	for _, want := range []string{"Examples:", "org2md capture.org", "-o - capture.org"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("--help output is missing %q, got:\n%s", want, stdout)
		}
	}
}

func TestTooManyArgumentsIsAnError(t *testing.T) {
	_, _, err := run(t, "", "a.org", "b.org")
	if err == nil {
		t.Fatal("Execute() = nil, want an argument error")
	}
	for _, want := range []string{"one org file", "--help"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want it to mention %q", err, want)
		}
	}
}

func TestConvertsBesideTheInput(t *testing.T) {
	input := orgFile(t)

	if _, _, err := run(t, "", input); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	got := readFile(t, strings.TrimSuffix(input, ".org")+".md")
	if got != sampleMarkdown {
		t.Errorf("converted file =\n%q\nwant\n%q", got, sampleMarkdown)
	}
}

func TestOutputToStdout(t *testing.T) {
	input := orgFile(t)

	stdout, _, err := run(t, "", "-o", "-", input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stdout != sampleMarkdown {
		t.Errorf("stdout =\n%q\nwant\n%q", stdout, sampleMarkdown)
	}
	if _, err := os.Stat(strings.TrimSuffix(input, ".org") + ".md"); !os.IsNotExist(err) {
		t.Error("writing to stdout also wrote a file")
	}
}

func TestOutputToPath(t *testing.T) {
	input := orgFile(t)
	dest := filepath.Join(filepath.Dir(input), "elsewhere.md")

	if _, _, err := run(t, "", "--output", dest, input); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := readFile(t, dest); got != sampleMarkdown {
		t.Errorf("converted file =\n%q\nwant\n%q", got, sampleMarkdown)
	}
}

func TestRefusesToOverwriteWithoutForce(t *testing.T) {
	input := orgFile(t)
	dest := strings.TrimSuffix(input, ".org") + ".md"
	if err := os.WriteFile(dest, []byte("old\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, _, err := run(t, "", input); err == nil {
		t.Fatal("Execute() error = nil, want an already-exists error")
	}
	if got := readFile(t, dest); got != "old\n" {
		t.Errorf("file = %q, want the original %q", got, "old\n")
	}

	if _, _, err := run(t, "", "--force", input); err != nil {
		t.Fatalf("Execute() with --force error = %v", err)
	}
	if got := readFile(t, dest); got != sampleMarkdown {
		t.Error("--force did not overwrite the file")
	}
}

func TestMissingInput(t *testing.T) {
	if _, _, err := run(t, "", filepath.Join(t.TempDir(), "absent.org")); err == nil {
		t.Fatal("Execute() error = nil, want a missing-file error")
	}
}

func TestNonOrgExtensionIsOnlyAWarning(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(input, []byte(sampleOrg), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, stderr, err := run(t, "", input)
	if err != nil {
		t.Fatalf("Execute() error = %v, want a warning only", err)
	}
	if !strings.Contains(stderr, "does not look like an org file") {
		t.Errorf("stderr = %q, want a note about the extension", stderr)
	}
	readFile(t, filepath.Join(dir, "note.md"))
}

func TestWarningsAreCountedAndDetailed(t *testing.T) {
	input := orgFile(t)

	_, quiet, err := run(t, "", "-o", "-", input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(quiet, "1 construct could not be converted") {
		t.Errorf("stderr = %q, want a count of the warnings", quiet)
	}
	if strings.Contains(quiet, "OPTIONS") {
		t.Errorf("stderr = %q, want details held back without --verbose", quiet)
	}

	_, loud, err := run(t, "", "--verbose", "-o", "-", input)
	if err != nil {
		t.Fatalf("Execute() with --verbose error = %v", err)
	}
	if !strings.Contains(loud, "#+OPTIONS") {
		t.Errorf("verbose stderr = %q, want the warning itself", loud)
	}
}

func TestWarningCountIsPluralised(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "many.org")
	src := "#+OPTIONS: toc:nil\n#+SUBTITLE: x\n\nBody.\n"
	if err := os.WriteFile(input, []byte(src), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, stderr, err := run(t, "", "-o", "-", input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stderr, "2 constructs could not be converted") {
		t.Errorf("stderr = %q, want a plural count", stderr)
	}
}

func TestMovesIntoTheNotesDirectory(t *testing.T) {
	tests := []struct {
		name      string
		stdin     string
		args      []string
		wantMoved bool
	}{
		{name: "answering yes", stdin: "y\n", wantMoved: true},
		{name: "answering no", stdin: "n\n"},
		{name: "--yes", args: []string{"--yes"}, wantMoved: true},
		{name: "--no-move", stdin: "y\n", args: []string{"--no-move"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := orgFile(t)
			notes := filepath.Join(filepath.Dir(input), "vault")
			t.Setenv("MDNOTES_DIR", notes)

			args := append(tt.args, input)
			if _, _, err := run(t, tt.stdin, args...); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			inboxed := filepath.Join(notes, "Inbox", "note.md")
			beside := strings.TrimSuffix(input, ".org") + ".md"

			if tt.wantMoved {
				if got := readFile(t, inboxed); got != sampleMarkdown {
					t.Error("the moved file does not hold the conversion")
				}
				if _, err := os.Stat(beside); !os.IsNotExist(err) {
					t.Error("the file was left beside the input after moving")
				}
				return
			}

			readFile(t, beside)
			if _, err := os.Stat(inboxed); !os.IsNotExist(err) {
				t.Error("the file was moved when it should not have been")
			}
		})
	}
}

func TestNoNotesDirectoryMeansNoPrompt(t *testing.T) {
	input := orgFile(t)
	t.Setenv("MDNOTES_DIR", "")

	_, stderr, err := run(t, "y\n", input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if strings.Contains(stderr, "Move") {
		t.Errorf("stderr = %q, want no prompt without a notes directory", stderr)
	}
}

// TestAnswerReaderDeclinesANonTerminal guards the reason term.IsTerminal is
// used: /dev/null is a character device, so the os.ModeCharDevice idiom reads
// it as a terminal and prints a prompt nothing can answer.
func TestAnswerReaderDeclinesANonTerminal(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("Open(%q) error = %v", os.DevNull, err)
	}
	defer f.Close()

	if got := answerReader(f); got != nil {
		t.Errorf("answerReader(%s) = %v, want nil", os.DevNull, got)
	}
}

// A reader that is not a file, such as a test's own buffer, is always askable.
func TestAnswerReaderAcceptsAPlainReader(t *testing.T) {
	in := strings.NewReader("y\n")
	if got := answerReader(in); got != in {
		t.Errorf("answerReader(buffer) = %v, want the reader itself", got)
	}
}
