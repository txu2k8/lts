package utils

import (
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/fatih/color"
	"github.com/minio/pkg/console"
)

var printMu sync.Mutex

// progress extender.
type ProgressBar struct {
	*pb.ProgressBar
}

// newProgressBar - instantiate a progress bar.
func NewProgressBar(total int64, units pb.Units) *ProgressBar {
	// Progress bar specific theme customization.
	console.SetColor("Bar", color.New(color.FgGreen, color.Bold))

	pgbar := ProgressBar{}

	// get the new original progress bar.
	bar := pb.New64(total)

	// Set new human friendly print units.
	bar.SetUnits(pb.U_DURATION)

	// Refresh rate for progress bar is set to 125 milliseconds.
	bar.SetRefreshRate(time.Millisecond * 125)

	// Do not print a newline by default handled, it is handled manually.
	bar.NotPrint = true

	// Show current speed is true.
	bar.ShowSpeed = false
	bar.ShowTimeLeft = false
	bar.ManualUpdate = true

	// Custom callback with colorized bar.
	bar.Callback = func(s string) {
		printMu.Lock()
		defer printMu.Unlock()
		console.Print(console.Colorize("Bar", "\r"+s+"\r"))
	}

	// Use different unicodes for Linux, OS X and Windows.
	switch runtime.GOOS {
	case "linux", "windows":
		// Need to add '\x00' as delimiter for unicode characters.
		bar.Format("┃\x00▓\x00█\x00░\x00┃")
	case "darwin":
		// Need to add '\x00' as delimiter for unicode characters.
		bar.Format(" \x00▓\x00 \x00░\x00 ")
	default:
		// Default to non unicode characters.
		bar.Format("[=> ]")
	}

	// Start the progress bar.
	if bar.Total > 0 {
		bar.Start()
	}

	// Copy for future
	pgbar.ProgressBar = bar

	// Return new progress bar here.
	return &pgbar
}

// Set caption.
func (p *ProgressBar) SetCaption(caption string) *ProgressBar {
	caption = FixateBarCaption(caption, GetFixedWidth(p.ProgressBar.GetWidth(), 18))
	p.ProgressBar.Prefix(caption)
	return p
}

func (p *ProgressBar) Set64(length int64) *ProgressBar {
	p.ProgressBar = p.ProgressBar.Set64(length)
	return p
}

func (p *ProgressBar) SetTotal(total int64) *ProgressBar {
	p.ProgressBar.Total = total
	return p
}

// fixateBarCaption - fancify bar caption based on the terminal width.
func FixateBarCaption(caption string, width int) string {
	switch {
	case len(caption) > width:
		// Trim caption to fit within the screen
		trimSize := len(caption) - width + 3
		if trimSize < len(caption) {
			caption = "..." + caption[trimSize:]
		}
	case len(caption) < width:
		caption += strings.Repeat(" ", width-len(caption))
	}
	return caption
}

// getFixedWidth - get a fixed width based for a given percentage.
func GetFixedWidth(width, percent int) int {
	return width * percent / 100
}
