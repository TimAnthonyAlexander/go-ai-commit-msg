package ui

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/schollz/progressbar/v3"
)

// LoadingSpinner creates a beautiful loading spinner
func NewLoadingSpinner(message string) *spinner.Spinner {
	if IsNoColor() {
		// Simple spinner for no-color environments
		s := spinner.New([]string{"|", "/", "-", "\\"}, 100*time.Millisecond)
		s.Suffix = " " + message
		return s
	}

	// Beautiful spinner with colors
	s := spinner.New([]string{
		"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
	}, 80*time.Millisecond)

	s.Suffix = InfoStyle.Render(" " + message)
	s.Color("cyan")

	return s
}

// StreamingSpinner creates a special spinner for streaming responses
func NewStreamingSpinner(message string) *StreamingSpinner {
	return &StreamingSpinner{
		message: message,
		dots:    0,
		maxDots: 3,
	}
}

// StreamingSpinner provides a custom streaming indicator
type StreamingSpinner struct {
	message string
	dots    int
	maxDots int
	started bool
}

// Start begins the streaming animation
func (s *StreamingSpinner) Start() {
	if !s.started {
		fmt.Print(InfoStyle.Render(s.message))
		s.started = true
	}
}

// Update adds a dot to the streaming animation
func (s *StreamingSpinner) Update() {
	if !s.started {
		s.Start()
	}

	if IsNoColor() {
		fmt.Print(".")
	} else {
		fmt.Print(MutedStyle.Render("●"))
	}

	s.dots++
	if s.dots > s.maxDots {
		s.dots = 0
	}
}

// Stop finishes the streaming animation
func (s *StreamingSpinner) Stop() {
	if s.started {
		fmt.Println() // New line
	}
}

// NewProgressBar creates a beautiful progress bar
func NewProgressBar(max int, description string) *progressbar.ProgressBar {
	if IsNoColor() {
		return progressbar.NewOptions(max,
			progressbar.OptionSetDescription(description),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetPredictTime(true),
			progressbar.OptionFullWidth(),
			progressbar.OptionThrottle(65*time.Millisecond),
		)
	}

	return progressbar.NewOptions(max,
		progressbar.OptionSetDescription(InfoStyle.Render(description)),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "▐",
			BarEnd:        "▌",
		}),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(false),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionThrottle(65*time.Millisecond),
	)
}

// AnimatedMessage creates an animated message display
type AnimatedMessage struct {
	frames   []string
	current  int
	ticker   *time.Ticker
	stopChan chan bool
	writer   io.Writer
	message  string
}

// NewAnimatedMessage creates a new animated message
func NewAnimatedMessage(message string, writer io.Writer) *AnimatedMessage {
	if writer == nil {
		writer = os.Stdout
	}

	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	if IsNoColor() {
		frames = []string{"|", "/", "-", "\\"}
	}

	return &AnimatedMessage{
		frames:   frames,
		current:  0,
		stopChan: make(chan bool),
		writer:   writer,
		message:  message,
	}
}

// Start begins the animation
func (a *AnimatedMessage) Start() {
	a.ticker = time.NewTicker(100 * time.Millisecond)

	go func() {
		for {
			select {
			case <-a.ticker.C:
				a.update()
			case <-a.stopChan:
				return
			}
		}
	}()
}

// Stop ends the animation
func (a *AnimatedMessage) Stop() {
	if a.ticker != nil {
		a.ticker.Stop()
		a.stopChan <- true
		fmt.Fprint(a.writer, "\r") // Clear the line
	}
}

// update refreshes the animation frame
func (a *AnimatedMessage) update() {
	frame := a.frames[a.current]
	if IsNoColor() {
		fmt.Fprintf(a.writer, "\r%s %s", frame, a.message)
	} else {
		fmt.Fprintf(a.writer, "\r%s %s",
			InfoStyle.Render(frame),
			BodyStyle.Render(a.message))
	}

	a.current = (a.current + 1) % len(a.frames)
}

// ShowSuccess displays a success message with animation
func ShowSuccess(message string) {
	if IsNoColor() {
		fmt.Printf("✓ %s\n", message)
	} else {
		fmt.Println(RenderSuccessBox(message))
	}
}

// ShowError displays an error message with animation
func ShowError(message string) {
	if IsNoColor() {
		fmt.Printf("✗ %s\n", message)
	} else {
		fmt.Println(RenderErrorBox(message))
	}
}

// ShowWarning displays a warning message with animation
func ShowWarning(message string) {
	if IsNoColor() {
		fmt.Printf("⚠ %s\n", message)
	} else {
		fmt.Println(RenderWarningBox(message))
	}
}

// ShowInfo displays an info message
func ShowInfo(message string) {
	if IsNoColor() {
		fmt.Printf("ℹ %s\n", message)
	} else {
		fmt.Println(InfoStyle.Render("ℹ ") + BodyStyle.Render(message))
	}
}
