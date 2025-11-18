package textwrap

import (
	"reflect"
	"testing"
)

func TestDetectCodeBlocks(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []Range
	}{
		{
			name:     "inline code block",
			text:     "this is `code` block",
			expected: []Range{{Start: 8, End: 14, Type: RangeTypeCodeBlock}},
		},
		{
			name:     "multiline code block",
			text:     "```\nfunc main() {\n}\n```",
			expected: []Range{{Start: 0, End: 23, Type: RangeTypeCodeBlock}},
		},
		{
			name:     "multiple inline code blocks",
			text:     "`foo` and `bar` here",
			expected: []Range{{Start: 0, End: 5, Type: RangeTypeCodeBlock}, {Start: 10, End: 15, Type: RangeTypeCodeBlock}},
		},
		{
			name:     "incomplete code block (opening only)",
			text:     "this is `incomplete",
			expected: []Range{},
		},
		{
			name:     "nested code block",
			text:     "```\n`nested`\n```",
			expected: []Range{{Start: 0, End: 16, Type: RangeTypeCodeBlock}},
		},
		{
			name:     "no code blocks",
			text:     "plain text here",
			expected: []Range{},
		},
		{
			name:     "empty string",
			text:     "",
			expected: []Range{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewSpecialFormatDetector()
			result := detector.DetectCodeBlocks(tt.text)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("DetectCodeBlocks() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetectURLs(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []Range
	}{
		{
			name:     "HTTP URL",
			text:     "Check this http://example.com for details",
			expected: []Range{{Start: 11, End: 29, Type: RangeTypeURL}},
		},
		{
			name:     "HTTPS URL",
			text:     "Visit https://example.com/path?query=1",
			expected: []Range{{Start: 6, End: 38, Type: RangeTypeURL}},
		},
		{
			name:     "multiple URLs",
			text:     "See http://a.com and https://b.com here",
			expected: []Range{{Start: 4, End: 16, Type: RangeTypeURL}, {Start: 21, End: 34, Type: RangeTypeURL}},
		},
		{
			name:     "no URLs",
			text:     "This is plain text without links",
			expected: []Range{},
		},
		{
			name:     "empty string",
			text:     "",
			expected: []Range{},
		},
		{
			name:     "URL with query parameters",
			text:     "Link: https://example.com/page?foo=bar&baz=qux",
			expected: []Range{{Start: 6, End: 46, Type: RangeTypeURL}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewSpecialFormatDetector()
			result := detector.DetectURLs(tt.text)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("DetectURLs() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetectQuotes(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []Range
	}{
		{
			name:     "single line quote",
			text:     "> This is a quote",
			expected: []Range{{Start: 0, End: 17, Type: RangeTypeQuote}},
		},
		{
			name: "multiline quotes",
			text: "> First line\n> Second line\n> Third line",
			expected: []Range{
				{Start: 0, End: 12, Type: RangeTypeQuote},
				{Start: 13, End: 26, Type: RangeTypeQuote},
				{Start: 27, End: 39, Type: RangeTypeQuote},
			},
		},
		{
			name:     "quote with extra spaces",
			text:     ">  Quote with extra space",
			expected: []Range{{Start: 0, End: 25, Type: RangeTypeQuote}},
		},
		{
			name:     "no quotes",
			text:     "This is not a quote",
			expected: []Range{},
		},
		{
			name:     "empty string",
			text:     "",
			expected: []Range{},
		},
		{
			name:     "mixed text with quote",
			text:     "Normal text\n> Quote line\nMore normal text",
			expected: []Range{{Start: 12, End: 24, Type: RangeTypeQuote}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewSpecialFormatDetector()
			result := detector.DetectQuotes(tt.text)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("DetectQuotes() = %v, want %v", result, tt.expected)
			}
		})
	}
}
