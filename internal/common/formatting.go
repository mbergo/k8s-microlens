package common

import (
	"fmt"
	"strings"
)

const (
	ColorRed    = "\033[0;31m"
	ColorGreen  = "\033[0;32m"
	ColorBlue   = "\033[0;34m"
	ColorYellow = "\033[1;33m"
	ColorCyan   = "\033[0;36m"
	ColorReset  = "\033[0m"
)

type Formatter struct {
	indent int
}

func NewFormatter() *Formatter {
	return &Formatter{
		indent: 0,
	}
}

func (f *Formatter) PrintHeader(text string) {
	fmt.Printf("%s%s%s\n", ColorGreen, text, ColorReset)
}

func (f *Formatter) PrintLine() {
	fmt.Println(strings.Repeat("-", 80))
}

func (f *Formatter) PrintSuccess(text string) {
	fmt.Printf("%s%s%s\n", ColorGreen, text, ColorReset)
}

func (f *Formatter) PrintResource(icon, resourceType, name string) {
	fmt.Printf("%s%s● %s/%s%s\n", f.getIndent(), ColorBlue, resourceType, name, ColorReset)
}

func (f *Formatter) PrintInfo(icon string, format string, a ...interface{}) {
	fmt.Printf("%s%sℹ %s%s\n", f.getIndent(), ColorCyan, fmt.Sprintf(format, a...), ColorReset)
}

func (f *Formatter) PrintStatus(status string, ok bool) {
	icon := "✓"
	color := ColorGreen
	if !ok {
		icon = "✗"
		color = ColorRed
	}
	fmt.Printf("%s%s%s %s%s\n", f.getIndent(), color, icon, status, ColorReset)
}

func (f *Formatter) PrintRelation(resourceType, name string, details ...string) {
	fmt.Printf("%s➜ %s/%s%s\n", f.getIndent(), resourceType, name, ColorReset)
	for _, detail := range details {
		fmt.Printf("%s  %s\n", f.getIndent(), detail)
	}
}

func (f *Formatter) Indent() {
	f.indent++
}

func (f *Formatter) Outdent() {
	if f.indent > 0 {
		f.indent--
	}
}

func (f *Formatter) getIndent() string {
	return strings.Repeat("    ", f.indent)
}
