// Package tui provides TUI (Text User Interface) components using Bubble Tea.
package tui

import (
	"strings"
	"testing"
)

func TestStringBuilderPool_Get(t *testing.T) {
	pool := NewStringBuilderPool()

	// Get a string builder
	sb := pool.Get()
	if sb == nil {
		t.Error("Expected non-nil string builder")
	}

	// Should be empty
	if sb.Len() != 0 {
		t.Error("Expected empty string builder")
	}
}

func TestStringBuilderPool_Put(t *testing.T) {
	pool := NewStringBuilderPool()

	// Get, use, and put back
	sb := pool.Get()
	sb.WriteString("test content")

	pool.Put(sb)

	// Get again - should be reset
	sb2 := pool.Get()
	if sb2.Len() != 0 {
		t.Error("Expected string builder to be reset after Put")
	}
}

func TestStringBuilderPool_Reuse(t *testing.T) {
	pool := NewStringBuilderPool()

	// Verify pool reuses builders
	sb1 := pool.Get()
	sb1.WriteString("first")
	pool.Put(sb1)

	sb2 := pool.Get()
	// Should be the same underlying builder (or at least from pool)
	if sb2.Len() != 0 {
		t.Error("Expected reused builder to be reset")
	}
}

func TestStringBuilderPool_ConcurrentAccess(t *testing.T) {
	pool := NewStringBuilderPool()
	done := make(chan bool)

	// Concurrent get/put operations
	for i := 0; i < 100; i++ {
		go func() {
			sb := pool.Get()
			sb.WriteString("concurrent test")
			pool.Put(sb)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}

func BenchmarkStringBuilderPool_WithPool(b *testing.B) {
	pool := NewStringBuilderPool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sb := pool.Get()
		sb.WriteString("benchmark test content with some length")
		_ = sb.String()
		pool.Put(sb)
	}
}

func BenchmarkStringBuilderPool_WithoutPool(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var sb strings.Builder
		sb.WriteString("benchmark test content with some length")
		_ = sb.String()
	}
}

func BenchmarkStringBuilderPool_LargeContent(b *testing.B) {
	pool := NewStringBuilderPool()
	content := strings.Repeat("x", 1000) // 1KB content

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sb := pool.Get()
		sb.WriteString(content)
		_ = sb.String()
		pool.Put(sb)
	}
}
