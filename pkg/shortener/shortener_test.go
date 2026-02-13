package shortener

import (
	"sync"
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "simple URL",
			url:  "https://google.com",
		},
		{
			name: "long URL",
			url:  "https://example.com/very/long/path/with/many/segments?param1=value1&param2=value2",
		},
		{
			name: "URL with special characters",
			url:  "https://example.com/path?query=value&special=@#$%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortCode := Generate(tt.url)
			if len(shortCode) != ShortURLLength {
				t.Errorf("Generate() returned length %d, want %d", len(shortCode), ShortURLLength)
			}

			if !Validate(shortCode) {
				t.Errorf("Generate() returned invalid short code: %s", shortCode)
			}
		})
	}
}

func TestGenerateIdempotency(t *testing.T) {
	url := "https://example.com/test"

	results := make([]string, 100)
	for i := 0; i < 100; i++ {
		results[i] = Generate(url)
	}

	first := results[0]
	for i, result := range results {
		if result != first {
			t.Errorf("Generate() is not idempotent: call %d returned %s, expected %s", i, result, first)
		}
	}
}

func TestGenerateDifferentURLs(t *testing.T) {
	url1 := "https://example.com/page1"
	url2 := "https://example.com/page2"

	short1 := Generate(url1)
	short2 := Generate(url2)

	if short1 == short2 {
		t.Errorf("Generate() returned same short code for different URLs: %s", short1)
	}
}

func TestGenerateConcurrency(t *testing.T) {
	url := "https://example.com/concurrent"
	expectedShort := Generate(url)

	var wg sync.WaitGroup
	results := make(chan string, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- Generate(url)
		}()
	}

	wg.Wait()
	close(results)

	for result := range results {
		if result != expectedShort {
			t.Errorf("Concurrent Generate() returned %s, expected %s", result, expectedShort)
		}
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		shortCode string
		want      bool
	}{
		{
			name:      "valid code",
			shortCode: "aBc123XY_z",
			want:      true,
		},
		{
			name:      "too short",
			shortCode: "abc123",
			want:      false,
		},
		{
			name:      "too long",
			shortCode: "abc123XYZ123",
			want:      false,
		},
		{
			name:      "invalid character -",
			shortCode: "abc123-XYZ",
			want:      false,
		},
		{
			name:      "invalid character @",
			shortCode: "abc123@XYZ",
			want:      false,
		},
		{
			name:      "all lowercase",
			shortCode: "abcdefghij",
			want:      true,
		},
		{
			name:      "all uppercase",
			shortCode: "ABCDEFGHIJ",
			want:      true,
		},
		{
			name:      "all numbers",
			shortCode: "0123456789",
			want:      true,
		},
		{
			name:      "with underscores",
			shortCode: "abc_123_XY",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Validate(tt.shortCode); got != tt.want {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
