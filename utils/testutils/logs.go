package testutils

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"testing"
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
	log.SetOutput(&out)
	return &out
}

func (c capturedLogs) Logs() []string {
	return c.stack
}

// CheckEqual compares logs ignoring date time in logged.
func (c capturedLogs) CheckEqual(refs []string, t *testing.T) {
	gots := c.Logs()
	for i, ref := range refs {
		g := gots[i][20:]
		if g != ref {
			t.Fatalf("expected \n%s\n got \n%s", ref, g)
		}
	}
}

func (c *capturedLogs) AssertNoLogs(t *testing.T) {
	l := c.Logs()
	if len(l) > 0 {
		t.Fatalf("expected no logs, got (%d): \n %s", len(l), strings.Join(l, "\n"))
	}
}

// IndentLogger enable to write debug message with a tree structure.
type IndentLogger struct {
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

// Line simply writes the message without changing the indentation.
func (il *IndentLogger) Line(format string, args ...interface{}) {
	fmt.Println(strings.Repeat(" ", il.level) + fmt.Sprintf(format, args...))
}
