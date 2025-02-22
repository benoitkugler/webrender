package testutils

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/benoitkugler/webrender/logger"
)

type capturedLogs struct {
	stack []string
}

func (c *capturedLogs) Write(p []byte) (int, error) {
	b := new(bytes.Buffer)
	i, err := b.Write(p)
	if err != nil {
		return i, err
	}
	c.stack = append(c.stack, strings.TrimSuffix(b.String(), "\n"))
	return i, nil
}

func CaptureLogs() *capturedLogs {
	out := capturedLogs{}
	logger.WarningLogger.SetOutput(&out)
	return &out
}

func (c capturedLogs) Logs() []string {
	return c.stack
}

// CheckEqual compares logs ignoring date time in logged.
func (c capturedLogs) CheckEqual(refs []string, t *testing.T) {
	t.Helper()

	const prefixLength = len("webrender.warning: ")
	gots := c.Logs()
	if len(gots) != len(refs) {
		t.Fatalf("expected %d logs, got %d", len(refs), len(gots))
	}
	for i, ref := range refs {
		g := gots[i][prefixLength:]
		if g != ref {
			t.Fatalf("expected \n%s\n got \n%s", ref, g)
		}
	}
}

func (c *capturedLogs) AssertNoLogs(t *testing.T) {
	t.Helper()

	l := c.Logs()
	if len(l) > 0 {
		t.Fatalf("expected no logs, got (%d): \n %s", len(l), strings.Join(l, "\n"))
	}
}

// IndentLogger enable to write debug message with a tree structure.
type IndentLogger struct {
	Color bool
	level int
}

// LineWithIndent prints the message with the given indent level, then increases it.
func (il *IndentLogger) LineWithIndent(format string, args ...interface{}) {
	il.Line(format, args...)
	il.level++
}

// LineWithDedent decreases the level, then write the message.
func (il *IndentLogger) LineWithDedent(format string, args ...interface{}) {
	il.level--
	il.Line(format, args...)
}

var reTag = regexp.MustCompile(`<(\S+)>`)

func colorTag(s string) string {
	return reTag.ReplaceAllString(s, "\033[1;34m<$1>\033[0m")
}

// Line simply writes the message without changing the indentation.
func (il *IndentLogger) Line(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	if il.Color {
		s = colorTag(s)
	}
	fmt.Println(strings.Repeat(" ", il.level) + s)
}
