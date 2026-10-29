package main

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var update = flag.Bool("update", false, "rewrite the golden file in testdata")

const realisticOrg = "testdata/realistic.org"
const realisticMarkdown = "testdata/realistic.md"

// build compiles the command the way a user would, so the end-to-end tests
// exercise the real binary rather than the package's internals.
func build(t *testing.T) string {
	t.Helper()

	bin := filepath.Join(t.TempDir(), "org2md")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build error = %v\n%s", err, out)
	}
	return bin
}

// TestEndToEndStdout converts a note holding every supported construct and
// compares the binary's own output with the golden file. Run `make golden` to
// rewrite it after an intended change, then read the diff.
func TestEndToEndStdout(t *testing.T) {
	if testing.Short() {
		t.Skip("builds the binary")
	}
	bin := build(t)

	cmd := exec.Command(bin, "--verbose", "-o", "-", realisticOrg)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("org2md error = %v\nstderr: %s", err, stderr.String())
	}

	if *update {
		if err := os.WriteFile(realisticMarkdown, []byte(stdout.String()), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		return
	}

	want, err := os.ReadFile(realisticMarkdown)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v (run `make golden` to create it)", realisticMarkdown, err)
	}
	if diff := cmp.Diff(string(want), stdout.String()); diff != "" {
		t.Errorf("converted output mismatch (-%s +got):\n%s", realisticMarkdown, diff)
	}

	// The only thing in the note that Markdown cannot carry is #+OPTIONS.
	if !strings.Contains(stderr.String(), "dropped #+OPTIONS") {
		t.Errorf("stderr = %q, want the dropped keyword reported", stderr.String())
	}
}

// TestEndToEndFilesIntoTheInbox walks the path a capture actually takes:
// convert beside the note, then move it into the vault.
func TestEndToEndFilesIntoTheInbox(t *testing.T) {
	if testing.Short() {
		t.Skip("builds the binary")
	}
	bin := build(t)

	source, err := os.ReadFile(realisticOrg)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", realisticOrg, err)
	}
	want, err := os.ReadFile(realisticMarkdown)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", realisticMarkdown, err)
	}

	dir := t.TempDir()
	input := filepath.Join(dir, "capture.org")
	if err := os.WriteFile(input, source, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	notes := filepath.Join(dir, "second-brain")

	cmd := exec.Command(bin, "--yes", "--verbose", input)
	cmd.Env = append(os.Environ(), "MDNOTES_DIR="+notes)
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("org2md error = %v\nstderr: %s", err, stderr.String())
	}

	inboxed := filepath.Join(notes, "Inbox", "capture.md")
	got, err := os.ReadFile(inboxed)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v; the note was not filed", inboxed, err)
	}
	if diff := cmp.Diff(string(want), string(got)); diff != "" {
		t.Errorf("filed note mismatch (-want +got):\n%s", diff)
	}

	if _, err := os.Stat(filepath.Join(dir, "capture.md")); !os.IsNotExist(err) {
		t.Error("the note was left beside the input after being filed")
	}
	if !strings.Contains(stderr.String(), inboxed) {
		t.Errorf("stderr = %q, want the filed path reported under --verbose", stderr.String())
	}
}
