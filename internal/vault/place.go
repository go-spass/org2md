// Package vault decides where converted Markdown goes and, when a notes
// directory is known, offers to file it there.
package vault

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// inbox is the subdirectory of the notes directory new notes land in.
const inbox = "Inbox"

// Placer writes converted Markdown to its destination. Everything it needs
// from the environment is a field, so its behaviour is decided by the caller
// rather than discovered here.
type Placer struct {
	// NotesDir is $MDNOTES_DIR, empty when unset.
	NotesDir string

	Force  bool
	Yes    bool
	NoMove bool

	Stdout io.Writer

	// Prompt is where the move question is written and Answers where the reply
	// is read. A nil Answers means there is nowhere to ask, so the move is
	// declined rather than assumed.
	Prompt  io.Writer
	Answers io.Reader
}

// Place writes content for the given input file and reports the path it ended
// up at, or an empty path when it went to stdout.
//
// An explicit output path is final: no prompt and no move. With no output path
// the Markdown is written beside the input, and is then offered to the notes
// directory if there is one.
func (p *Placer) Place(input, output, content string) (string, error) {
	if output == "-" {
		_, err := io.WriteString(p.Stdout, content)
		return "", err
	}
	if output != "" {
		if err := p.write(output, content); err != nil {
			return "", err
		}
		return output, nil
	}

	beside := MarkdownPath(input)
	if err := p.write(beside, content); err != nil {
		return "", err
	}

	target, ok := p.inboxPath(beside)
	if !ok || !p.confirm(beside, target) {
		return beside, nil
	}
	if err := p.move(beside, target); err != nil {
		// The conversion already succeeded; report the path that holds it.
		return beside, err
	}

	return target, nil
}

// MarkdownPath is the Markdown file that sits beside an org file.
func MarkdownPath(input string) string {
	dir, base := filepath.Split(input)
	return filepath.Join(dir, strings.TrimSuffix(base, filepath.Ext(base))+".md")
}

func (p *Placer) inboxPath(written string) (string, bool) {
	if p.NotesDir == "" || p.NoMove {
		return "", false
	}
	return filepath.Join(p.NotesDir, inbox, filepath.Base(written)), true
}

func (p *Placer) confirm(from, to string) bool {
	if p.Yes {
		return true
	}
	if p.Answers == nil {
		return false
	}

	fmt.Fprintf(p.Prompt, "Move %s to %s? [y/N] ", filepath.Base(from), filepath.Dir(to))
	line, _ := bufio.NewReader(p.Answers).ReadString('\n')

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	}
	return false
}

func (p *Placer) write(path, content string) error {
	if err := p.checkVacant(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func (p *Placer) checkVacant(path string) error {
	if p.Force {
		return nil
	}
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists; pass --force to overwrite", path)
	}
	return nil
}

func (p *Placer) move(from, to string) error {
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return err
	}
	if err := p.checkVacant(to); err != nil {
		return err
	}

	if err := os.Rename(from, to); err == nil {
		return nil
	}
	// Rename cannot cross filesystems, and a notes directory may well be on
	// another volume, so fall back to copying.
	if err := copyFile(from, to); err != nil {
		return err
	}
	return os.Remove(from)
}

func copyFile(from, to string) error {
	src, err := os.Open(from)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(to)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return err
	}
	return dst.Close()
}
