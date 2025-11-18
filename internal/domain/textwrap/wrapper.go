package textwrap

import (
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/reflow/wordwrap"
)

// TextWrapOptions represents text wrapping configuration for domain layer (external format independent).
type TextWrapOptions struct {
	Enabled               bool
	MaxLineWidth          int
	BreakAtCJKPunctuation bool
}

// TextWrapper provides pure business logic for text wrapping.
type TextWrapper struct {
	detector *SpecialFormatDetector
}

// NewTextWrapper creates a new instance of TextWrapper.
func NewTextWrapper() *TextWrapper {
	return &TextWrapper{
		detector: NewSpecialFormatDetector(),
	}
}

// WrapText wraps text according to specified width.
// It considers special formats (code blocks, quotes, URLs) and preserves original line breaks.
//
// Parameters:
//   - text: Text to wrap
//   - width: Terminal width (character count)
//   - opts: Text wrapping options
//
// Returns:
//   - Wrapped text
//   - Error (if width is 0 or less)
func (tw *TextWrapper) WrapText(text string, width int, opts TextWrapOptions) (string, error) {
	// Precondition: width > 0
	if width <= 0 {
		return "", fmt.Errorf("invalid width: %d", width)
	}

	// Return text as-is if wrapping is disabled
	if !opts.Enabled {
		return text, nil
	}

	// Return empty string for empty text
	if text == "" {
		return "", nil
	}

	// Detect special formats in the entire text
	specialRanges := tw.detector.DetectSpecialFormats(text)

	// Split text by newlines to preserve original line breaks
	lines := strings.Split(text, "\n")
	wrappedLines := make([]string, 0, len(lines))

	// Track position in the original text
	pos := 0

	// Wrap each line independently
	for i, line := range lines {
		lineStart := pos
		lineEnd := pos + len(line)

		// Check if this line is inside a code block
		inCodeBlock := tw.isInRange(lineStart, lineEnd, specialRanges, RangeTypeCodeBlock)

		// Check if this line is a quote
		isQuote := strings.HasPrefix(strings.TrimSpace(line), ">")

		if inCodeBlock {
			// Don't wrap code blocks
			wrappedLines = append(wrappedLines, line)
		} else if isQuote {
			// Wrap quote with prefix preserved
			wrapped, err := tw.wrapQuote(line, width, opts)
			if err != nil {
				return "", err
			}
			wrappedLines = append(wrappedLines, wrapped)
		} else {
			// Normal wrapping with URL detection
			wrapped, err := tw.wrapWithURLDetection(line, width, opts, specialRanges, lineStart)
			if err != nil {
				return "", err
			}
			wrappedLines = append(wrappedLines, wrapped)
		}

		// Update position (add line length + newline)
		pos = lineEnd
		if i < len(lines)-1 {
			pos++ // for \n
		}
	}

	// Join wrapped lines with newlines
	return strings.Join(wrappedLines, "\n"), nil
}

// isInRange checks if a text range overlaps with any special format range
func (tw *TextWrapper) isInRange(start, end int, ranges []Range, rangeType RangeType) bool {
	for _, r := range ranges {
		if r.Type == rangeType && start < r.End && end > r.Start {
			return true
		}
	}
	return false
}

// wrapQuote wraps a quoted line, preserving the quote prefix
func (tw *TextWrapper) wrapQuote(line string, width int, opts TextWrapOptions) (string, error) {
	// Extract quote prefix and content
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, ">") {
		return tw.WrapLine(line, width, opts)
	}

	// Remove quote prefix
	content := strings.TrimSpace(trimmed[1:])

	// Calculate available width for content (subtract "> " prefix)
	quotePrefix := "> "
	availableWidth := width - len(quotePrefix)
	if availableWidth < 5 {
		// Width too small, return as-is
		return line, nil
	}

	// Wrap the content
	wrapped, err := tw.WrapLine(content, availableWidth, opts)
	if err != nil {
		return "", err
	}

	// Add quote prefix to each wrapped line
	wrappedLines := strings.Split(wrapped, "\n")
	for i := range wrappedLines {
		wrappedLines[i] = quotePrefix + wrappedLines[i]
	}

	return strings.Join(wrappedLines, "\n"), nil
}

// wrapWithURLDetection wraps text with URL detection
func (tw *TextWrapper) wrapWithURLDetection(line string, width int, opts TextWrapOptions, specialRanges []Range, lineStart int) (string, error) {
	// For simplicity, use normal wrapping
	// URL detection is already handled by SpecialFormatDetector
	// URLs in ranges should not be split by word boundary
	return tw.WrapLine(line, width, opts)
}

// WrapLine wraps a single line of text (internal helper function).
func (tw *TextWrapper) WrapLine(line string, width int, opts TextWrapOptions) (string, error) {
	// Precondition: width > 0
	if width <= 0 {
		return "", fmt.Errorf("invalid width: %d", width)
	}

	// Return empty string for empty line
	if line == "" {
		return "", nil
	}

	// Calculate display width considering CJK characters and ANSI codes
	displayWidth := ansi.PrintableRuneWidth(line)
	if displayWidth <= width {
		return line, nil
	}

	// Check if line contains any spaces
	hasSpaces := strings.Contains(line, " ")

	// If no spaces or very short width, use forced wrapping
	if !hasSpaces || width < 5 {
		return tw.forcedWrapLine(line, width)
	}

	// Use wordwrap for intelligent word boundary wrapping with ANSI awareness
	ww := wordwrap.NewWriter(width)

	// Add CJK punctuation as breakpoints if enabled
	if opts.BreakAtCJKPunctuation {
		ww.Breakpoints = append(ww.Breakpoints, '、', '。')
	}

	// Write the line to the word wrapper
	_, err := ww.Write([]byte(line))
	if err != nil {
		// If wordwrap fails, fall back to forced wrapping
		return tw.forcedWrapLine(line, width)
	}

	// Close the writer and get the result
	if err := ww.Close(); err != nil {
		// If close fails, fall back to forced wrapping
		return tw.forcedWrapLine(line, width)
	}

	// Get wrapped result
	wrapped := ww.String()

	// Remove trailing newline if present (wordwrap adds it)
	wrapped = strings.TrimSuffix(wrapped, "\n")

	return wrapped, nil
}

// forcedWrapLine performs forced wrapping when wordwrap is not suitable.
// It considers CJK character widths and preserves ANSI escape sequences.
func (tw *TextWrapper) forcedWrapLine(line string, width int) (string, error) {
	// Calculate display width using ANSI-aware function
	displayWidth := ansi.PrintableRuneWidth(line)
	if displayWidth <= width {
		return line, nil
	}

	// Forced wrapping: split line into chunks considering display width
	var result strings.Builder
	var currentWidth int
	var currentChunk strings.Builder

	// Track ANSI sequences
	inAnsi := false
	var ansiBuffer strings.Builder

	for _, r := range line {
		// Check if this is an ANSI escape sequence start
		if r == '\033' {
			inAnsi = true
			ansiBuffer.Reset()
			ansiBuffer.WriteRune(r)
			continue
		}

		// Continue building ANSI sequence
		if inAnsi {
			ansiBuffer.WriteRune(r)
			// ANSI sequences end with a letter
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inAnsi = false
				currentChunk.WriteString(ansiBuffer.String())
			}
			continue
		}

		// Calculate display width of this rune
		runeDisplayWidth := runewidth.RuneWidth(r)

		// Check if adding this rune exceeds width
		if currentWidth+runeDisplayWidth > width && currentWidth > 0 {
			// Write current chunk to result
			if result.Len() > 0 {
				result.WriteRune('\n')
			}
			result.WriteString(currentChunk.String())

			// Start new chunk
			currentChunk.Reset()
			currentWidth = 0
		}

		// Add rune to current chunk
		currentChunk.WriteRune(r)
		currentWidth += runeDisplayWidth
	}

	// Write final chunk
	if currentChunk.Len() > 0 {
		if result.Len() > 0 {
			result.WriteRune('\n')
		}
		result.WriteString(currentChunk.String())
	}

	return result.String(), nil
}
