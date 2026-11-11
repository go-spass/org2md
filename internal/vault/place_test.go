package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixture lays out a temporary input file and an inbox directory.
type fixture struct {
	dir      string
	input    string
	inboxDir string
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	dir := t.TempDir()
	input := filepath.Join(dir, "note.org")
	if err := os.WriteFile(input, []byte("* Heading\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	// The inbox name ends in "Inbox" on purpose: if Placer ever appends a
	// subdirectory again, the target becomes .../Inbox/Inbox/note.md and the
	// move assertions below catch it.
	inboxDir := filepath.Join(dir, "vault", "Inbox")
	// org2md files into a vault it does not create, so the inbox has to exist
	// before a move is offered. Tests about an inbox that does not exist build
	// their own path instead of using this one.
	if err := os.MkdirAll(inboxDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	return fixture{dir: dir, input: input, inboxDir: inboxDir}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return string(b)
}

func TestMarkdownPath(t *testing.T) {
	tests := []struct{ in, want string }{
		{"note.org", "note.md"},
		{"/a/b/note.org", "/a/b/note.md"},
		{"/a/b/note.txt", "/a/b/note.md"},
		{"/a/b/note", "/a/b/note.md"},
		{"/a/b/note.name.org", "/a/b/note.name.md"},
	}
	for _, tt := range tests {
		if got := MarkdownPath(tt.in); got != tt.want {
			t.Errorf("MarkdownPath(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestPlaceStdout(t *testing.T) {
	f := newFixture(t)
	var out strings.Builder
	p := &Placer{Stdout: &out}

	path, err := p.Place(f.input, "-", "# Heading\n")
	if err != nil {
		t.Fatalf("Place() error = %v", err)
	}
	if path != "" {
		t.Errorf("Place() path = %q, want empty for stdout", path)
	}
	if out.String() != "# Heading\n" {
		t.Errorf("stdout = %q, want %q", out.String(), "# Heading\n")
	}
	if _, err := os.Stat(MarkdownPath(f.input)); !os.IsNotExist(err) {
		t.Error("Place() wrote a file as well as stdout")
	}
}

func TestPlaceExplicitOutput(t *testing.T) {
	f := newFixture(t)
	want := filepath.Join(f.dir, "elsewhere.md")
	p := &Placer{InboxDir: f.inboxDir}

	path, err := p.Place(f.input, want, "# Heading\n")
	if err != nil {
		t.Fatalf("Place() error = %v", err)
	}
	if path != want {
		t.Errorf("Place() path = %q, want %q", path, want)
	}
	if got := readFile(t, want); got != "# Heading\n" {
		t.Errorf("file = %q, want %q", got, "# Heading\n")
	}
	// An explicit destination is final: no prompt and no move, even with an
	// inbox set.
	if _, err := os.Stat(filepath.Join(f.inboxDir, "elsewhere.md")); !os.IsNotExist(err) {
		t.Error("Place() moved a file given an explicit output path")
	}
}

func TestPlaceBesideInput(t *testing.T) {
	f := newFixture(t)
	p := &Placer{}

	path, err := p.Place(f.input, "", "# Heading\n")
	if err != nil {
		t.Fatalf("Place() error = %v", err)
	}
	if want := MarkdownPath(f.input); path != want {
		t.Errorf("Place() path = %q, want %q", path, want)
	}
	readFile(t, MarkdownPath(f.input))
}

func TestPlaceRefusesToOverwrite(t *testing.T) {
	f := newFixture(t)
	dest := MarkdownPath(f.input)
	if err := os.WriteFile(dest, []byte("old\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	p := &Placer{}
	if _, err := p.Place(f.input, "", "new\n"); err == nil {
		t.Fatal("Place() error = nil, want an already-exists error")
	}
	if got := readFile(t, dest); got != "old\n" {
		t.Errorf("file = %q, want the original %q", got, "old\n")
	}

	p.Force = true
	if _, err := p.Place(f.input, "", "new\n"); err != nil {
		t.Fatalf("Place() with Force error = %v", err)
	}
	if got := readFile(t, dest); got != "new\n" {
		t.Errorf("file = %q, want %q", got, "new\n")
	}
}

func TestPlaceMovesIntoInbox(t *testing.T) {
	tests := []struct {
		name       string
		placer     Placer
		answers    string
		wantMoved  bool
		wantPrompt bool
	}{
		{name: "yes accepts without asking", placer: Placer{Yes: true}, wantMoved: true},
		{name: "answering y moves", answers: "y\n", wantMoved: true, wantPrompt: true},
		{name: "answering yes moves", answers: "YES\n", wantMoved: true, wantPrompt: true},
		{name: "answering n declines", answers: "n\n", wantPrompt: true},
		{name: "an empty answer declines", answers: "\n", wantPrompt: true},
		{name: "no-move never asks", placer: Placer{NoMove: true}, answers: "y\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			var prompt strings.Builder

			p := tt.placer
			p.InboxDir = f.inboxDir
			p.Prompt = &prompt
			if tt.answers != "" {
				p.Answers = strings.NewReader(tt.answers)
			}

			path, err := p.Place(f.input, "", "# Heading\n")
			if err != nil {
				t.Fatalf("Place() error = %v", err)
			}

			inbox := filepath.Join(f.inboxDir, "note.md")
			beside := MarkdownPath(f.input)

			if tt.wantMoved {
				if path != inbox {
					t.Errorf("Place() path = %q, want %q", path, inbox)
				}
				readFile(t, inbox)
				if _, err := os.Stat(beside); !os.IsNotExist(err) {
					t.Error("Place() left the file beside the input after moving it")
				}
			} else {
				if path != beside {
					t.Errorf("Place() path = %q, want %q", path, beside)
				}
				readFile(t, beside)
				if _, err := os.Stat(inbox); !os.IsNotExist(err) {
					t.Error("Place() moved the file when it should not have")
				}
			}

			if gotPrompt := prompt.Len() > 0; gotPrompt != tt.wantPrompt {
				t.Errorf("prompt written = %v, want %v (%q)", gotPrompt, tt.wantPrompt, prompt.String())
			}
		})
	}
}

// TestPlaceMovesIntoAnInboxWithSpaces covers the destination this tool is
// actually pointed at: a PARA vault's "00 - Inbox", whose name has spaces and
// is not called "Inbox" at all.
func TestPlaceMovesIntoAnInboxWithSpaces(t *testing.T) {
	f := newFixture(t)
	inboxDir := filepath.Join(f.dir, "second-brain", "00 - Inbox")
	if err := os.MkdirAll(inboxDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	p := &Placer{InboxDir: inboxDir, Yes: true}

	path, err := p.Place(f.input, "", "# Heading\n")
	if err != nil {
		t.Fatalf("Place() error = %v", err)
	}

	want := filepath.Join(inboxDir, "note.md")
	if path != want {
		t.Errorf("Place() path = %q, want %q", path, want)
	}
	if got := readFile(t, want); got != "# Heading\n" {
		t.Errorf("file = %q, want %q", got, "# Heading\n")
	}
	if _, err := os.Stat(MarkdownPath(f.input)); !os.IsNotExist(err) {
		t.Error("Place() left the file beside the input after moving it")
	}
}

// A configured inbox that org2md cannot file into is fatal, and fatal means
// nothing was written: the check runs before the beside-the-input write, so a
// typo in the variable costs a rerun rather than a stray file and an invented
// directory tree.
func TestPlaceRefusesAMissingInbox(t *testing.T) {
	f := newFixture(t)
	missing := filepath.Join(f.dir, "second-brain", "00 - Inbx")
	p := &Placer{InboxDir: missing, Yes: true}

	path, err := p.Place(f.input, "", "# Heading\n")
	if err == nil {
		t.Fatal("Place() error = nil, want a missing-inbox error")
	}
	if !strings.Contains(err.Error(), missing) {
		t.Errorf("Place() error = %q, want it to name %q", err, missing)
	}
	if path != "" {
		t.Errorf("Place() path = %q, want empty when nothing was written", path)
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Errorf("Place() created %q; org2md files into a vault, it does not reshape one", missing)
	}
}

func TestPlaceRefusesANonDirectoryInbox(t *testing.T) {
	f := newFixture(t)
	blocked := filepath.Join(f.dir, "blocked")
	if err := os.WriteFile(blocked, []byte("not a directory\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	p := &Placer{InboxDir: blocked, Yes: true}
	path, err := p.Place(f.input, "", "# Heading\n")
	if err == nil {
		t.Fatal("Place() error = nil, want a not-a-directory error")
	}
	if !strings.Contains(err.Error(), blocked) {
		t.Errorf("Place() error = %q, want it to name %q", err, blocked)
	}
	if path != "" {
		t.Errorf("Place() path = %q, want empty when nothing was written", path)
	}
	if got := readFile(t, blocked); got != "not a directory\n" {
		t.Errorf("file at the inbox path = %q, want it untouched", got)
	}
}

func TestPlaceLeavesNothingBehindWhenTheInboxIsMissing(t *testing.T) {
	f := newFixture(t)
	p := &Placer{InboxDir: filepath.Join(f.dir, "absent"), Yes: true}

	if _, err := p.Place(f.input, "", "# Heading\n"); err == nil {
		t.Fatal("Place() error = nil, want a missing-inbox error")
	}
	if _, err := os.Stat(MarkdownPath(f.input)); !os.IsNotExist(err) {
		t.Error("Place() left the Markdown beside the input; a fatal condition writes nothing")
	}
}

func TestPlaceDoesNotAskWhenTheInboxIsUnusable(t *testing.T) {
	f := newFixture(t)
	var prompt strings.Builder
	p := &Placer{
		InboxDir: filepath.Join(f.dir, "absent"),
		Prompt:   &prompt,
		Answers:  strings.NewReader("y\n"),
	}

	if _, err := p.Place(f.input, "", "# Heading\n"); err == nil {
		t.Fatal("Place() error = nil, want a missing-inbox error")
	}
	if prompt.Len() != 0 {
		t.Errorf("prompt = %q, want no question asked for a destination that cannot be honored", prompt.String())
	}
}

// The check is gated on a move actually being in play, so a run that was never
// going to touch the inbox is unaffected by a bad value.
func TestPlaceIgnoresAnUnusableInboxWithoutAMove(t *testing.T) {
	tests := []struct {
		name   string
		output string
		noMove bool
	}{
		{name: "explicit output", output: "chosen.md"},
		{name: "stdout", output: "-"},
		{name: "no-move", noMove: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			var out strings.Builder
			output := tt.output
			if output != "" && output != "-" {
				output = filepath.Join(f.dir, output)
			}

			p := &Placer{
				InboxDir: filepath.Join(f.dir, "absent"),
				NoMove:   tt.noMove,
				Yes:      true,
				Stdout:   &out,
			}
			if _, err := p.Place(f.input, output, "# Heading\n"); err != nil {
				t.Fatalf("Place() error = %v, want an unusable inbox to be irrelevant here", err)
			}
		})
	}
}

func TestPlaceCannotAskWithoutAnswers(t *testing.T) {
	f := newFixture(t)
	var prompt strings.Builder
	p := &Placer{InboxDir: f.inboxDir, Prompt: &prompt}

	path, err := p.Place(f.input, "", "# Heading\n")
	if err != nil {
		t.Fatalf("Place() error = %v", err)
	}
	if want := MarkdownPath(f.input); path != want {
		t.Errorf("Place() path = %q, want %q", path, want)
	}
	if prompt.Len() != 0 {
		t.Errorf("prompt = %q, want nothing asked when there is nowhere to read an answer", prompt.String())
	}
}

func TestPlaceRefusesToOverwriteInInbox(t *testing.T) {
	f := newFixture(t)
	inbox := f.inboxDir
	existing := filepath.Join(inbox, "note.md")
	if err := os.WriteFile(existing, []byte("old\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	p := &Placer{InboxDir: f.inboxDir, Yes: true}
	path, err := p.Place(f.input, "", "new\n")
	if err == nil {
		t.Fatal("Place() error = nil, want an already-exists error")
	}

	// The conversion is not thrown away just because the move failed.
	beside := MarkdownPath(f.input)
	if path != beside {
		t.Errorf("Place() path = %q, want the file it did write, %q", path, beside)
	}
	if got := readFile(t, beside); got != "new\n" {
		t.Errorf("file beside input = %q, want %q", got, "new\n")
	}
	if got := readFile(t, existing); got != "old\n" {
		t.Errorf("inbox file = %q, want the original %q", got, "old\n")
	}
}

// TestPlaceExplicitOutputRefusesToOverwrite covers the explicit-output branch
// of the vacancy check, which the beside-the-input tests do not reach.
func TestPlaceExplicitOutputRefusesToOverwrite(t *testing.T) {
	f := newFixture(t)
	dest := filepath.Join(f.dir, "taken.md")
	if err := os.WriteFile(dest, []byte("old\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	p := &Placer{}
	if _, err := p.Place(f.input, dest, "new\n"); err == nil {
		t.Fatal("Place() error = nil, want an already-exists error")
	}
	if got := readFile(t, dest); got != "old\n" {
		t.Errorf("file = %q, want the original %q", got, "old\n")
	}
}

// TestMoveByCopy exercises the fallback used when rename cannot cross a
// filesystem boundary. The copy is driven directly, since provoking a real
// EXDEV in a test would need a second volume.
func TestMoveByCopy(t *testing.T) {
	dir := t.TempDir()
	from := filepath.Join(dir, "from.md")
	to := filepath.Join(dir, "sub", "to.md")

	if err := os.WriteFile(from, []byte("contents\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := copyFile(from, to); err != nil {
		t.Fatalf("copyFile() error = %v", err)
	}

	if got := readFile(t, to); got != "contents\n" {
		t.Errorf("copy = %q, want %q", got, "contents\n")
	}
	if got := readFile(t, from); got != "contents\n" {
		t.Errorf("source = %q, want it left in place for the caller to remove", got)
	}
}

func TestCopyFileReportsAMissingSource(t *testing.T) {
	dir := t.TempDir()
	if err := copyFile(filepath.Join(dir, "absent.md"), filepath.Join(dir, "to.md")); err == nil {
		t.Fatal("copyFile() error = nil, want a missing-source error")
	}
}

func TestCopyFileReportsAnUnwritableTarget(t *testing.T) {
	dir := t.TempDir()
	from := filepath.Join(dir, "from.md")
	if err := os.WriteFile(from, []byte("contents\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	locked := filepath.Join(dir, "locked")
	if err := os.Mkdir(locked, 0o555); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	t.Cleanup(func() { os.Chmod(locked, 0o755) })

	if err := copyFile(from, filepath.Join(locked, "to.md")); err == nil {
		t.Fatal("copyFile() error = nil, want a permission error")
	}
}
