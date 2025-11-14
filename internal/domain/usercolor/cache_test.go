package usercolor

import (
	"testing"
)

func TestNewUserColorCache(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		wantSize int
	}{
		{
			name:     "create cache with size 10",
			size:     10,
			wantSize: 10,
		},
		{
			name:     "create cache with size 100",
			size:     100,
			wantSize: 100,
		},
		{
			name:     "create cache with size 500 (default)",
			size:     500,
			wantSize: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewUserColorCache(tt.size)
			if cache == nil {
				t.Fatal("NewUserColorCache() returned nil")
			}

			// Verify cache is empty initially
			if cache.Len() != 0 {
				t.Errorf("NewUserColorCache() initial length = %v, want 0", cache.Len())
			}
		})
	}
}

func TestCache_GetSet(t *testing.T) {
	cache := NewUserColorCache(10)

	redLight := Color{R: 255, G: 100, B: 100}
	redDark := Color{R: 200, G: 50, B: 50}
	redColor := AdaptiveColor{Light: redLight, Dark: redDark}

	blueLight := Color{R: 100, G: 100, B: 255}
	blueDark := Color{R: 50, G: 50, B: 200}
	blueColor := AdaptiveColor{Light: blueLight, Dark: blueDark}

	// Test Get on empty cache
	t.Run("get from empty cache", func(t *testing.T) {
		_, ok := cache.Get("U12345")
		if ok {
			t.Error("Get() on empty cache should return ok=false")
		}
	})

	// Test Set and Get
	t.Run("set and get single entry", func(t *testing.T) {
		cache.Set("U12345", redColor)

		got, ok := cache.Get("U12345")
		if !ok {
			t.Fatal("Get() should return ok=true for existing entry")
		}

		if got.Light != redLight || got.Dark != redDark {
			t.Errorf("Get() returned wrong color: got %+v, want %+v", got, redColor)
		}

		if cache.Len() != 1 {
			t.Errorf("cache.Len() = %v, want 1", cache.Len())
		}
	})

	// Test multiple Set/Get operations
	t.Run("set and get multiple entries", func(t *testing.T) {
		cache.Clear()
		cache.Set("U11111", redColor)
		cache.Set("U22222", blueColor)

		// Get first entry
		got1, ok1 := cache.Get("U11111")
		if !ok1 {
			t.Fatal("Get() should return ok=true for U11111")
		}
		if got1.Light != redLight {
			t.Errorf("Get(U11111) returned wrong color")
		}

		// Get second entry
		got2, ok2 := cache.Get("U22222")
		if !ok2 {
			t.Fatal("Get() should return ok=true for U22222")
		}
		if got2.Light != blueLight {
			t.Errorf("Get(U22222) returned wrong color")
		}

		if cache.Len() != 2 {
			t.Errorf("cache.Len() = %v, want 2", cache.Len())
		}
	})

	// Test overwrite existing entry
	t.Run("overwrite existing entry", func(t *testing.T) {
		cache.Clear()
		cache.Set("U12345", redColor)
		cache.Set("U12345", blueColor) // Overwrite with blue

		got, ok := cache.Get("U12345")
		if !ok {
			t.Fatal("Get() should return ok=true")
		}
		if got.Light != blueLight {
			t.Errorf("Get() returned old color, want updated color")
		}

		if cache.Len() != 1 {
			t.Errorf("cache.Len() = %v, want 1 (no duplication)", cache.Len())
		}
	})
}

func TestCache_LRUEviction(t *testing.T) {
	// Create cache with size 3
	cache := NewUserColorCache(3)

	color1 := AdaptiveColor{Light: Color{R: 1, G: 0, B: 0}, Dark: Color{R: 1, G: 0, B: 0}}
	color2 := AdaptiveColor{Light: Color{R: 2, G: 0, B: 0}, Dark: Color{R: 2, G: 0, B: 0}}
	color3 := AdaptiveColor{Light: Color{R: 3, G: 0, B: 0}, Dark: Color{R: 3, G: 0, B: 0}}
	color4 := AdaptiveColor{Light: Color{R: 4, G: 0, B: 0}, Dark: Color{R: 4, G: 0, B: 0}}

	// Fill cache to capacity
	cache.Set("U1", color1)
	cache.Set("U2", color2)
	cache.Set("U3", color3)

	if cache.Len() != 3 {
		t.Fatalf("cache.Len() = %v, want 3", cache.Len())
	}

	// Add 4th entry - should evict U1 (least recently used)
	cache.Set("U4", color4)

	if cache.Len() != 3 {
		t.Errorf("cache.Len() = %v, want 3 (should maintain size limit)", cache.Len())
	}

	// U1 should be evicted
	_, ok := cache.Get("U1")
	if ok {
		t.Error("U1 should have been evicted")
	}

	// U2, U3, U4 should still exist
	_, ok2 := cache.Get("U2")
	if !ok2 {
		t.Error("U2 should still exist")
	}

	_, ok3 := cache.Get("U3")
	if !ok3 {
		t.Error("U3 should still exist")
	}

	_, ok4 := cache.Get("U4")
	if !ok4 {
		t.Error("U4 should exist (just added)")
	}
}

func TestCache_LRUEvictionWithAccess(t *testing.T) {
	// Test that accessing an entry makes it recently used
	cache := NewUserColorCache(3)

	color1 := AdaptiveColor{Light: Color{R: 1, G: 0, B: 0}, Dark: Color{R: 1, G: 0, B: 0}}
	color2 := AdaptiveColor{Light: Color{R: 2, G: 0, B: 0}, Dark: Color{R: 2, G: 0, B: 0}}
	color3 := AdaptiveColor{Light: Color{R: 3, G: 0, B: 0}, Dark: Color{R: 3, G: 0, B: 0}}
	color4 := AdaptiveColor{Light: Color{R: 4, G: 0, B: 0}, Dark: Color{R: 4, G: 0, B: 0}}

	cache.Set("U1", color1)
	cache.Set("U2", color2)
	cache.Set("U3", color3)

	// Access U1 to make it recently used
	cache.Get("U1")

	// Add U4 - should evict U2 (now least recently used)
	cache.Set("U4", color4)

	// U1 should still exist (we just accessed it)
	_, ok1 := cache.Get("U1")
	if !ok1 {
		t.Error("U1 should still exist (was recently accessed)")
	}

	// U2 should be evicted
	_, ok2 := cache.Get("U2")
	if ok2 {
		t.Error("U2 should have been evicted (least recently used)")
	}

	// U3 and U4 should exist
	_, ok3 := cache.Get("U3")
	if !ok3 {
		t.Error("U3 should still exist")
	}

	_, ok4 := cache.Get("U4")
	if !ok4 {
		t.Error("U4 should exist")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewUserColorCache(10)

	color1 := AdaptiveColor{Light: Color{R: 1, G: 0, B: 0}, Dark: Color{R: 1, G: 0, B: 0}}
	color2 := AdaptiveColor{Light: Color{R: 2, G: 0, B: 0}, Dark: Color{R: 2, G: 0, B: 0}}

	// Add entries
	cache.Set("U1", color1)
	cache.Set("U2", color2)

	if cache.Len() != 2 {
		t.Fatalf("cache.Len() = %v, want 2", cache.Len())
	}

	// Clear cache
	cache.Clear()

	// Verify cache is empty
	if cache.Len() != 0 {
		t.Errorf("cache.Len() after Clear() = %v, want 0", cache.Len())
	}

	// Verify entries don't exist
	_, ok1 := cache.Get("U1")
	if ok1 {
		t.Error("U1 should not exist after Clear()")
	}

	_, ok2 := cache.Get("U2")
	if ok2 {
		t.Error("U2 should not exist after Clear()")
	}
}

func TestCache_Len(t *testing.T) {
	cache := NewUserColorCache(10)

	color := AdaptiveColor{Light: Color{R: 1, G: 0, B: 0}, Dark: Color{R: 1, G: 0, B: 0}}

	// Initial length
	if cache.Len() != 0 {
		t.Errorf("initial Len() = %v, want 0", cache.Len())
	}

	// Add entries and check length
	for i := 1; i <= 5; i++ {
		cache.Set("U"+string(rune('0'+i)), color)
		if cache.Len() != i {
			t.Errorf("Len() after %d additions = %v, want %d", i, cache.Len(), i)
		}
	}

	// Clear and check length
	cache.Clear()
	if cache.Len() != 0 {
		t.Errorf("Len() after Clear() = %v, want 0", cache.Len())
	}
}
