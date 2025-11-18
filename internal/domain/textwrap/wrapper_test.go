package textwrap

import (
	"strings"
	"testing"
	"time"
)

// TestNewTextWrapper tests NewTextWrapper constructor
func TestNewTextWrapper(t *testing.T) {
	tw := NewTextWrapper()
	if tw == nil {
		t.Fatal("NewTextWrapper() returned nil")
	}
	if tw.detector == nil {
		t.Error("TextWrapper.detector is nil")
	}
}

// TestWrapText_InvalidWidth tests width validation
func TestWrapText_InvalidWidth(t *testing.T) {
	tw := NewTextWrapper()

	tests := []struct {
		name  string
		width int
	}{
		{"zero width", 0},
		{"negative width", -1},
		{"large negative width", -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := TextWrapOptions{Enabled: true}
			_, err := tw.WrapText("test text", tt.width, opts)
			if err == nil {
				t.Errorf("WrapText() with width=%d should return error, got nil", tt.width)
			}
			if !strings.Contains(err.Error(), "invalid width") {
				t.Errorf("WrapText() error = %v, want error containing 'invalid width'", err)
			}
		})
	}
}

// TestWrapText_EmptyText tests empty text handling
func TestWrapText_EmptyText(t *testing.T) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true}

	result, err := tw.WrapText("", 80, opts)
	if err != nil {
		t.Errorf("WrapText() with empty text returned error: %v", err)
	}
	if result != "" {
		t.Errorf("WrapText() with empty text = %q, want empty string", result)
	}
}

// TestWrapText_DisabledWrapping tests that wrapping can be disabled
func TestWrapText_DisabledWrapping(t *testing.T) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: false}
	text := "This is a very long text that would normally be wrapped but wrapping is disabled"

	result, err := tw.WrapText(text, 20, opts)
	if err != nil {
		t.Errorf("WrapText() with disabled wrapping returned error: %v", err)
	}
	if result != text {
		t.Errorf("WrapText() with disabled wrapping = %q, want original text %q", result, text)
	}
}

// TestWrapText_ValidWidth tests that valid width is accepted
func TestWrapText_ValidWidth(t *testing.T) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true}

	tests := []struct {
		name  string
		width int
	}{
		{"minimum width", 1},
		{"small width", 10},
		{"medium width", 80},
		{"large width", 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := "test"
			_, err := tw.WrapText(text, tt.width, opts)
			if err != nil {
				t.Errorf("WrapText() with valid width=%d returned error: %v", tt.width, err)
			}
		})
	}
}

// TestTextWrapOptions_Validation tests TextWrapOptions validation
func TestTextWrapOptions_Validation(t *testing.T) {
	tw := NewTextWrapper()

	tests := []struct {
		name string
		opts TextWrapOptions
		text string
	}{
		{
			name: "enabled with MaxLineWidth=0",
			opts: TextWrapOptions{Enabled: true, MaxLineWidth: 0},
			text: "test text",
		},
		{
			name: "enabled with MaxLineWidth=80",
			opts: TextWrapOptions{Enabled: true, MaxLineWidth: 80},
			text: "test text",
		},
		{
			name: "disabled",
			opts: TextWrapOptions{Enabled: false},
			text: "test text",
		},
		{
			name: "BreakAtCJKPunctuation enabled",
			opts: TextWrapOptions{Enabled: true, BreakAtCJKPunctuation: true},
			text: "これはテストです。",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tw.WrapText(tt.text, 80, tt.opts)
			if err != nil {
				t.Errorf("WrapText() with opts=%+v returned error: %v", tt.opts, err)
			}
		})
	}
}

// TestWrapText_ForcedWrapping tests forced wrapping for long words
func TestWrapText_ForcedWrapping(t *testing.T) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true}

	tests := []struct {
		name     string
		text     string
		width    int
		expected string
	}{
		{
			name:     "word longer than width",
			text:     "verylongwordthatexceedswidth",
			width:    10,
			expected: "verylongwo\nrdthatexce\nedswidth",
		},
		{
			name:     "exactly width",
			text:     "exactwidth",
			width:    10,
			expected: "exactwidth",
		},
		{
			name:     "URL longer than width",
			text:     "https://example.com/very/long/path/that/exceeds/terminal/width",
			width:    20,
			expected: "https://example.com/\nvery/long/path/that/\nexceeds/terminal/wid\nth",
		},
		{
			name:     "single character per line",
			text:     "abc",
			width:    1,
			expected: "a\nb\nc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tw.WrapText(tt.text, tt.width, opts)
			if err != nil {
				t.Fatalf("WrapText() returned error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("WrapText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestWrapText_NewlinePreservation tests that original newlines are preserved
func TestWrapText_NewlinePreservation(t *testing.T) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true}

	tests := []struct {
		name     string
		text     string
		width    int
		expected string
	}{
		{
			name:     "single newline",
			text:     "line one\nline two",
			width:    80,
			expected: "line one\nline two",
		},
		{
			name:     "multiple consecutive newlines",
			text:     "line one\n\n\nline two",
			width:    80,
			expected: "line one\n\n\nline two",
		},
		{
			name:     "newline at start",
			text:     "\nline one",
			width:    80,
			expected: "\nline one",
		},
		{
			name:     "newline at end",
			text:     "line one\n",
			width:    80,
			expected: "line one\n",
		},
		{
			name:     "long lines with newlines",
			text:     "this is a very long line that needs wrapping\nthis is another line",
			width:    20,
			expected: "this is a very long\nline that needs\nwrapping\nthis is another line",
		},
		{
			name:     "each line wrapped independently",
			text:     "first line that is long\nsecond line also long",
			width:    15,
			expected: "first line that\nis long\nsecond line\nalso long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tw.WrapText(tt.text, tt.width, opts)
			if err != nil {
				t.Fatalf("WrapText() returned error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("WrapText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestWrapText_CombinedNewlineAndForced tests combination of newline preservation and forced wrapping
func TestWrapText_CombinedNewlineAndForced(t *testing.T) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true}

	tests := []struct {
		name     string
		text     string
		width    int
		expected string
	}{
		{
			name:     "long word with newline",
			text:     "verylongword\nnormal",
			width:    5,
			expected: "veryl\nongwo\nrd\nnorma\nl",
		},
		{
			name:     "multiple long words with newlines",
			text:     "firstlongword\nsecondlongword\nthird",
			width:    8,
			expected: "firstlon\ngword\nsecondlo\nngword\nthird",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tw.WrapText(tt.text, tt.width, opts)
			if err != nil {
				t.Fatalf("WrapText() returned error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("WrapText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestWrapText_WordBoundary tests wrapping at word boundaries
func TestWrapText_WordBoundary(t *testing.T) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true}

	tests := []struct {
		name     string
		text     string
		width    int
		expected string
	}{
		{
			name:     "simple sentence wrapped at space",
			text:     "hello world this is a test",
			width:    15,
			expected: "hello world\nthis is a test",
		},
		{
			name:     "wrap between words",
			text:     "the quick brown fox jumps over the lazy dog",
			width:    20,
			expected: "the quick brown fox\njumps over the lazy\ndog",
		},
		{
			name:     "multiple spaces preserved as single",
			text:     "hello    world",
			width:    20,
			expected: "hello    world",
		},
		{
			name:     "text exactly at width",
			text:     "exactly twenty ch",
			width:    17,
			expected: "exactly twenty ch",
		},
		{
			name:     "single word shorter than width",
			text:     "short",
			width:    10,
			expected: "short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tw.WrapText(tt.text, tt.width, opts)
			if err != nil {
				t.Fatalf("WrapText() returned error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("WrapText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestWrapText_CJKPunctuation tests wrapping at CJK punctuation marks
func TestWrapText_CJKPunctuation(t *testing.T) {
	tw := NewTextWrapper()

	tests := []struct {
		name                  string
		text                  string
		width                 int
		breakAtCJKPunctuation bool
		expected              string
	}{
		{
			name:                  "mixed English and Japanese with space",
			text:                  "This is English text これは日本語",
			width:                 20,
			breakAtCJKPunctuation: true,
			expected:              "This is English text\nこれは日本語",
		},
		{
			name:                  "mixed with CJK punctuation",
			text:                  "Hello world、日本語 text",
			width:                 15,
			breakAtCJKPunctuation: true,
			expected:              "Hello world、\n日本語 text",
		},
		{
			name:                  "English with CJK punct disabled",
			text:                  "Hello world、more text here",
			width:                 20,
			breakAtCJKPunctuation: false,
			expected:              "Hello world、more\ntext here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := TextWrapOptions{
				Enabled:               true,
				BreakAtCJKPunctuation: tt.breakAtCJKPunctuation,
			}
			result, err := tw.WrapText(tt.text, tt.width, opts)
			if err != nil {
				t.Fatalf("WrapText() returned error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("WrapText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestWrapText_Performance tests that wrapping 1000 chars completes within 1ms
func TestWrapText_Performance(t *testing.T) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true}

	// Generate 1000 character text
	text := strings.Repeat("Lorem ipsum dolor sit amet, consectetur adipiscing elit. ", 20)
	if len(text) < 1000 {
		text += strings.Repeat("x", 1000-len(text))
	}
	text = text[:1000]

	start := time.Now()
	_, err := tw.WrapText(text, 80, opts)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("WrapText() returned error: %v", err)
	}

	// Check if duration is less than 1ms
	if duration > time.Millisecond {
		t.Errorf("WrapText() took %v, want < 1ms", duration)
	} else {
		t.Logf("WrapText() completed in %v (requirement: < 1ms)", duration)
	}
}

// TestWrapText_CodeBlocks tests that code blocks are not wrapped
func TestWrapText_CodeBlocks(t *testing.T) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true}

	tests := []struct {
		name     string
		text     string
		width    int
		expected string
	}{
		{
			name:     "inline code block preserved",
			text:     "This is `inline code that should not wrap` in text",
			width:    20,
			expected: "This is `inline code that should not wrap` in text",
		},
		{
			name:     "multiline code block preserved",
			text:     "Text before\n```\ncode block line 1\ncode block line 2\n```\nText after",
			width:    15,
			expected: "Text before\n```\ncode block line 1\ncode block line 2\n```\nText after",
		},
		{
			name:     "code block with long line",
			text:     "Start ```verylongcodelinethatexceedswidth``` end",
			width:    10,
			expected: "Start ```verylongcodelinethatexceedswidth``` end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tw.WrapText(tt.text, tt.width, opts)
			if err != nil {
				t.Fatalf("WrapText() returned error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("WrapText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestWrapText_Quotes tests that quotes are wrapped with prefix preserved
func TestWrapText_Quotes(t *testing.T) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true}

	tests := []struct {
		name     string
		text     string
		width    int
		expected string
	}{
		{
			name:     "single line quote",
			text:     "> This is a quoted line",
			width:    20,
			expected: "> This is a quoted\n> line",
		},
		{
			name:     "long quote wrapped",
			text:     "> This is a very long quoted line that needs to be wrapped",
			width:    25,
			expected: "> This is a very long\n> quoted line that needs\n> to be wrapped",
		},
		{
			name:     "multiple quote lines",
			text:     "> Line one\n> Line two that is longer",
			width:    15,
			expected: "> Line one\n> Line two that\n> is longer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tw.WrapText(tt.text, tt.width, opts)
			if err != nil {
				t.Fatalf("WrapText() returned error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("WrapText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestWrapText_URLs tests that URLs are not wrapped
func TestWrapText_URLs(t *testing.T) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true}

	tests := []struct {
		name     string
		text     string
		width    int
		expected string
	}{
		{
			name:     "URL preserved in text",
			text:     "Check https://example.com/path for info",
			width:    20,
			expected: "Check\nhttps://example.com/path\nfor info",
		},
		{
			name:     "very long URL preserved",
			text:     "Visit https://example.com/very/long/path/that/exceeds/width here",
			width:    15,
			expected: "Visit\nhttps://example.com/very/long/path/that/exceeds/width\nhere",
		},
		{
			name:     "multiple URLs",
			text:     "Sites: https://a.com and https://b.com are good",
			width:    20,
			expected: "Sites: https://a.com\nand https://b.com\nare good",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tw.WrapText(tt.text, tt.width, opts)
			if err != nil {
				t.Fatalf("WrapText() returned error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("WrapText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestWrapText_ANSIPreservation tests that ANSI escape sequences are preserved during wrapping
func TestWrapText_ANSIPreservation(t *testing.T) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true}

	tests := []struct {
		name     string
		text     string
		width    int
		expected string
	}{
		{
			name:     "ANSI color codes preserved",
			text:     "\033[31mRed text\033[0m and \033[32mgreen text\033[0m here",
			width:    20,
			expected: "\033[31mRed text\033[0m and \033[32mgreen\ntext\033[0m here",
		},
		{
			name:     "ANSI bold and color preserved",
			text:     "\033[1m\033[34mBold blue text that is long\033[0m",
			width:    15,
			expected: "\033[1m\033[34mBold blue text\nthat is long\033[0m",
		},
		{
			name:     "Multiple ANSI codes in short text",
			text:     "\033[31mA\033[0m \033[32mB\033[0m \033[33mC\033[0m",
			width:    10,
			expected: "\033[31mA\033[0m \033[32mB\033[0m \033[33mC\033[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tw.WrapText(tt.text, tt.width, opts)
			if err != nil {
				t.Fatalf("WrapText() returned error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("WrapText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestWrapText_CJKCharacterWidth tests that CJK characters are correctly measured for wrapping
func TestWrapText_CJKCharacterWidth(t *testing.T) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true}

	tests := []struct {
		name     string
		text     string
		width    int
		expected string
	}{
		{
			name:     "Japanese full-width characters",
			text:     "あいうえおかきくけこ",
			width:    10,
			expected: "あいうえお\nかきくけこ",
		},
		{
			name:     "Chinese characters",
			text:     "你好世界这是测试",
			width:    8,
			expected: "你好世界\n这是测试",
		},
		{
			name:     "Mixed ASCII and Japanese",
			text:     "Hello あいうえお World",
			width:    15,
			expected: "Hello\nあいうえお\nWorld",
		},
		{
			name:     "Korean characters",
			text:     "안녕하세요 세계",
			width:    12,
			expected: "안녕하세요\n세계",
		},
		{
			name:     "Full-width numbers",
			text:     "１２３４５６７８９０",
			width:    10,
			expected: "１２３４５\n６７８９０",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tw.WrapText(tt.text, tt.width, opts)
			if err != nil {
				t.Fatalf("WrapText() returned error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("WrapText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestWrapText_CombinedANSIAndCJK tests wrapping with both ANSI codes and CJK characters
func TestWrapText_CombinedANSIAndCJK(t *testing.T) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true}

	tests := []struct {
		name     string
		text     string
		width    int
		expected string
	}{
		{
			name:     "ANSI colored Japanese text",
			text:     "\033[31m日本語のテキスト\033[0m",
			width:    10,
			expected: "\033[31m日本語のテ\nキスト\033[0m",
		},
		{
			name:     "Mixed colored English and Japanese",
			text:     "\033[32mHello\033[0m こんにちは \033[34mWorld\033[0m",
			width:    15,
			expected: "\033[32mHello\033[0m\nこんにちは\n\033[34mWorld\033[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tw.WrapText(tt.text, tt.width, opts)
			if err != nil {
				t.Fatalf("WrapText() returned error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("WrapText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// BenchmarkWrapText_1000Chars benchmarks wrapping 1000 characters
// Requirement: 1000文字のメッセージの折り返し計算時間が1ms以内
func BenchmarkWrapText_1000Chars(b *testing.B) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true}

	// Generate 1000 character text
	text := strings.Repeat("Lorem ipsum dolor sit amet, consectetur adipiscing elit. ", 20)
	if len(text) < 1000 {
		text += strings.Repeat("x", 1000-len(text))
	}
	text = text[:1000]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tw.WrapText(text, 80, opts)
		if err != nil {
			b.Fatalf("WrapText() returned error: %v", err)
		}
	}
}

// BenchmarkWrapText_ShortText benchmarks wrapping short text (baseline)
func BenchmarkWrapText_ShortText(b *testing.B) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true}
	text := "Hello world, this is a short message"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tw.WrapText(text, 80, opts)
		if err != nil {
			b.Fatalf("WrapText() returned error: %v", err)
		}
	}
}

// BenchmarkWrapText_WithNewlines benchmarks wrapping text with newlines
func BenchmarkWrapText_WithNewlines(b *testing.B) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true}
	text := "Line one\nLine two with some more text\nLine three\nLine four has even more content"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tw.WrapText(text, 40, opts)
		if err != nil {
			b.Fatalf("WrapText() returned error: %v", err)
		}
	}
}

// BenchmarkWrapText_WithANSI benchmarks wrapping text with ANSI escape sequences
func BenchmarkWrapText_WithANSI(b *testing.B) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true}
	text := "\033[31mRed text\033[0m and \033[32mgreen text\033[0m here with more content to wrap around the terminal width"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tw.WrapText(text, 40, opts)
		if err != nil {
			b.Fatalf("WrapText() returned error: %v", err)
		}
	}
}

// BenchmarkWrapText_WithCJK benchmarks wrapping CJK text
func BenchmarkWrapText_WithCJK(b *testing.B) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true, BreakAtCJKPunctuation: true}
	text := "日本語のテキストです。これはパフォーマンステストのためのサンプルテキストです。複数の文を含んでいます。"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tw.WrapText(text, 40, opts)
		if err != nil {
			b.Fatalf("WrapText() returned error: %v", err)
		}
	}
}

// BenchmarkWrapText_WithCodeBlocks benchmarks wrapping text with code blocks
func BenchmarkWrapText_WithCodeBlocks(b *testing.B) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true}
	text := "Here is some code: `func example() { return true }` and more text after the code block"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tw.WrapText(text, 40, opts)
		if err != nil {
			b.Fatalf("WrapText() returned error: %v", err)
		}
	}
}

// BenchmarkWrapText_WithURLs benchmarks wrapping text with URLs
func BenchmarkWrapText_WithURLs(b *testing.B) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: true}
	text := "Check https://example.com/very/long/path/that/might/need/handling for more info on this topic"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tw.WrapText(text, 40, opts)
		if err != nil {
			b.Fatalf("WrapText() returned error: %v", err)
		}
	}
}

// BenchmarkWrapText_Disabled benchmarks with wrapping disabled (baseline)
func BenchmarkWrapText_Disabled(b *testing.B) {
	tw := NewTextWrapper()
	opts := TextWrapOptions{Enabled: false}
	text := "This is a long text that would normally be wrapped but wrapping is disabled for this benchmark test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tw.WrapText(text, 40, opts)
		if err != nil {
			b.Fatalf("WrapText() returned error: %v", err)
		}
	}
}
