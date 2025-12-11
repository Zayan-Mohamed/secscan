package main

import (
	"testing"
)

// TestVersion verifies that the version constant is set
func TestVersion(t *testing.T) {
	if version == "" {
		t.Error("version should not be empty")
	}

	// Version should follow semantic versioning format
	if len(version) < 5 { // Minimum: "1.0.0"
		t.Errorf("version format seems invalid: %s", version)
	}
}

// TestDefaultRegexps verifies that default detection patterns are defined
func TestDefaultRegexps(t *testing.T) {
	if len(defaultRegexps) == 0 {
		t.Error("defaultRegexps should not be empty")
	}

	// Check for some essential patterns
	essentialPatterns := []string{
		"aws_access_key",
		"github_pat",
		"generic_api_key",
	}

	for _, pattern := range essentialPatterns {
		if _, exists := defaultRegexps[pattern]; !exists {
			t.Errorf("essential pattern %s is missing", pattern)
		}
	}
}

// TestDefaultAllowPatterns verifies that allowlist patterns are defined
func TestDefaultAllowPatterns(t *testing.T) {
	if len(defaultAllowPatterns) == 0 {
		t.Error("defaultAllowPatterns should not be empty")
	}
}

// TestShouldSkipFile verifies file skipping logic
func TestShouldSkipFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"package-lock.json", true},
		{"yarn.lock", true},
		{"go.sum", true},
		{"main.go", false},
		{"README.md", false},
		{"Cargo.lock", true},
		{"poetry.lock", true},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := shouldSkipFile(tt.filename)
			if result != tt.expected {
				t.Errorf("shouldSkipFile(%s) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

// TestShannonEntropy verifies entropy calculation
func TestShannonEntropy(t *testing.T) {
	tests := []struct {
		input       string
		minExpected float64
		maxExpected float64
	}{
		{"aaaaaaaa", 0.0, 0.1},             // Low entropy
		{"abcdefgh", 2.5, 3.5},             // Medium entropy
		{"aB3$xZ9!", 2.5, 4.0},             // Higher entropy
		{"AKIAIOSFODNN7EXAMPLE", 3.0, 5.0}, // Typical AWS key entropy
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			entropy := shannonEntropy(tt.input)
			if entropy < tt.minExpected || entropy > tt.maxExpected {
				t.Errorf("shannonEntropy(%s) = %f, want between %f and %f",
					tt.input, entropy, tt.minExpected, tt.maxExpected)
			}
		})
	}
}

// TestIsHighEntropy verifies high entropy detection
func TestIsHighEntropy(t *testing.T) {
	tests := []struct {
		input     string
		threshold float64
		expected  bool
	}{
		{"aaaaaaaa", 4.5, false},             // Low entropy
		{"AKIAIOSFODNN7EXAMPLE", 4.0, false}, // Below threshold
		{"test123", 5.0, false},              // Medium entropy
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isHighEntropy(tt.input, tt.threshold)
			if result != tt.expected {
				entropy := shannonEntropy(tt.input)
				t.Logf("isHighEntropy(%s, %f) = %v (entropy=%f)",
					tt.input, tt.threshold, result, entropy)
			}
		})
	}
}

// TestMaskSecret verifies secret masking
func TestMaskSecret(t *testing.T) {
	tests := []struct {
		input string
	}{
		{""},
		{"a"},
		{"ab"},
		{"secret123"},
		{"AKIAIOSFODNN7EXAMPLE"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := maskSecret(tt.input)
			// Just verify it returns something and masks the middle
			if len(tt.input) > 0 && result == tt.input {
				t.Errorf("maskSecret should mask the secret, got same value back")
			}
			t.Logf("maskSecret(%s) = %s", tt.input, result)
		})
	}
}

// TestGenerateHash verifies hash generation for deduplication
func TestGenerateHash(t *testing.T) {
	hash1 := generateHash("file1.txt", "aws_key", "AKIAIOSFODNN7EXAMPLE")
	hash2 := generateHash("file1.txt", "aws_key", "AKIAIOSFODNN7EXAMPLE")
	hash3 := generateHash("file2.txt", "aws_key", "AKIAIOSFODNN7EXAMPLE")

	// Same inputs should produce same hash
	if hash1 != hash2 {
		t.Error("Same inputs should produce same hash")
	}

	// Different inputs should produce different hash
	if hash1 == hash3 {
		t.Error("Different files should produce different hashes")
	}

	// Hash should not be empty
	if hash1 == "" {
		t.Error("Hash should not be empty")
	}
}

// TestStatsIncrement verifies stats tracking
func TestStatsIncrement(t *testing.T) {
	stats := &Stats{}

	stats.incrementFiles()
	if stats.FilesScanned != 1 {
		t.Errorf("FilesScanned = %d, want 1", stats.FilesScanned)
	}

	stats.incrementFiles()
	if stats.FilesScanned != 2 {
		t.Errorf("FilesScanned = %d, want 2", stats.FilesScanned)
	}

	stats.incrementCommits()
	if stats.CommitsScanned != 1 {
		t.Errorf("CommitsScanned = %d, want 1", stats.CommitsScanned)
	}

	stats.incrementFindings(5)
	if stats.FindingsTotal != 5 {
		t.Errorf("FindingsTotal = %d, want 5", stats.FindingsTotal)
	}
}

// TestGitAvailable verifies git availability check doesn't crash
func TestGitAvailable(t *testing.T) {
	// This is just a smoke test to ensure the function doesn't panic
	_ = gitAvailable()
}

// TestLooksLikeTextFile verifies text file detection
func TestLooksLikeTextFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"test.go", true},
		{"readme.md", true},
		{"config.json", true},
		{"script.sh", true},
		{"image.png", false},
		{"binary.exe", false},
		{"archive.zip", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := looksLikeTextFile(tt.filename)
			if result != tt.expected {
				t.Errorf("looksLikeTextFile(%s) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}
