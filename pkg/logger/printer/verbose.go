package logprinter

import (
	"fmt"
	"os"
	"strings"
)

var verbose bool

func init() {
	v := strings.ToLower(os.Getenv("TIUP_VERBOSE"))
	verbose = v == "1" || v == "enable"
}

// Verbose logs verbose messages
func Verbose(format string, args ...any) {
	if !verbose {
		return
	}
	fmt.Fprintln(stderr, "Verbose:", fmt.Sprintf(format, args...))
}

// Verbose logs verbose messages
func (l *Logger) Verbose(format string, args ...any) {
	if !verbose {
		return
	}
	fmt.Fprintln(l.stderr, "Verbose:", fmt.Sprintf(format, args...))
}
