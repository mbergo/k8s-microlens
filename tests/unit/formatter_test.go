package unit

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/mbergo/k8s-microlens/internal/common"
)

// captureOutput captures stdout for testing print functions
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestFormatter(t *testing.T) {
	formatter := common.NewFormatter()

	t.Run("PrintHeader", func(t *testing.T) {
		output := captureOutput(func() {
			formatter.PrintHeader("Test Header")
		})
		if !strings.Contains(output, "Test Header") {
			t.Errorf("Expected output to contain 'Test Header', got %s", output)
		}
	})

	t.Run("PrintLine", func(t *testing.T) {
		output := captureOutput(func() {
			formatter.PrintLine()
		})
		if len(output) < 80 {
			t.Errorf("Expected line length to be at least 80 characters, got %d", len(output))
		}
	})

	t.Run("PrintResource", func(t *testing.T) {
		output := captureOutput(func() {
			formatter.PrintResource("├──", "Pod", "test-pod")
		})
		expected := "Pod/test-pod"
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', got %s", expected, output)
		}
	})

	t.Run("Indentation", func(t *testing.T) {
		formatter.Indent()
		output := captureOutput(func() {
			formatter.PrintInfo("", "Test Info")
		})
		if !strings.HasPrefix(strings.TrimLeft(output, "\033["), "    ") {
			t.Errorf("Expected output to be indented with 4 spaces, got %s", output)
		}
		formatter.Outdent()
	})
}
