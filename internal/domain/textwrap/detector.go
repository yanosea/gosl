package textwrap

import (
	"regexp"
	"sort"
)

// RangeType represents the type of a range.
type RangeType int

const (
	RangeTypeCodeBlock RangeType = iota
	RangeTypeQuote
	RangeTypeURL
)

// Range represents a range within text.
type Range struct {
	Start int // Start position (byte offset)
	End   int // End position (byte offset)
	Type  RangeType
}

// SpecialFormatDetector detects special formats in text.
type SpecialFormatDetector struct{}

// NewSpecialFormatDetector creates a new instance of SpecialFormatDetector.
func NewSpecialFormatDetector() *SpecialFormatDetector {
	return &SpecialFormatDetector{}
}

// DetectCodeBlocks detects code blocks in text.
func (d *SpecialFormatDetector) DetectCodeBlocks(text string) []Range {
	ranges := []Range{}

	// Detect multiline code blocks (higher priority)
	multilinePattern := regexp.MustCompile("```[\\s\\S]*?```")
	multilineMatches := multilinePattern.FindAllStringIndex(text, -1)
	for _, match := range multilineMatches {
		ranges = append(ranges, Range{
			Start: match[0],
			End:   match[1],
			Type:  RangeTypeCodeBlock,
		})
	}

	// Detect inline code blocks
	inlinePattern := regexp.MustCompile("`([^`]+)`")
	inlineMatches := inlinePattern.FindAllStringIndex(text, -1)
	for _, match := range inlineMatches {
		// Check if it doesn't overlap with existing multiline code blocks
		overlaps := false
		for _, existingRange := range ranges {
			if match[0] >= existingRange.Start && match[1] <= existingRange.End {
				overlaps = true
				break
			}
		}
		if !overlaps {
			ranges = append(ranges, Range{
				Start: match[0],
				End:   match[1],
				Type:  RangeTypeCodeBlock,
			})
		}
	}

	// Sort by start position
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].Start < ranges[j].Start
	})

	return ranges
}

// DetectSpecialFormats detects all special formats in text.
// Priority: CodeBlock > URL > Quote
func (d *SpecialFormatDetector) DetectSpecialFormats(text string) []Range {
	// Currently only code blocks are implemented
	return d.DetectCodeBlocks(text)
}

// DetectQuotes detects quote blocks (lines starting with '>').
func (d *SpecialFormatDetector) DetectQuotes(text string) []Range {
	ranges := []Range{}

	// Pattern: line starting with '>' followed by optional whitespace
	quotePattern := regexp.MustCompile(`(?m)^>\s*.*$`)
	matches := quotePattern.FindAllStringIndex(text, -1)

	for _, match := range matches {
		ranges = append(ranges, Range{
			Start: match[0],
			End:   match[1],
			Type:  RangeTypeQuote,
		})
	}

	return ranges
}

// DetectURLs detects URLs in text.
func (d *SpecialFormatDetector) DetectURLs(text string) []Range {
	ranges := []Range{}

	// Reuse the existing highlightURLs() pattern
	urlPattern := regexp.MustCompile(`https?://[a-zA-Z0-9\-._~:/?#\[\]@!$&'()*+,;=%]+`)
	matches := urlPattern.FindAllStringIndex(text, -1)

	for _, match := range matches {
		ranges = append(ranges, Range{
			Start: match[0],
			End:   match[1],
			Type:  RangeTypeURL,
		})
	}

	return ranges
}
