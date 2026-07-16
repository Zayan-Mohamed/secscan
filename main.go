// secscan - Enhanced Go CLI secret scanner
// Version: 2.2.4
// Author: Zayan-Mohamed (itsm.zayan@gmail.com)
// License: MIT
//
// Installation:
//
//	make install                         # build and install to /usr/local/bin (recommended)
//
// Usage examples:
//
//	secscan -root .                      # scan working tree + git history
//	secscan -root . -history=false       # scan only current files
//	secscan -root . -json report.json    # output JSON report
//	secscan -root . -config .secscan.toml  # use custom config
//	secscan -root . -entropy 4.5         # adjust entropy threshold
//	secscan -root . -min-confidence 0.8  # only report high-confidence findings
//	secscan -root . -verbose             # show detailed output
//	secscan -root . -respect-gitignore=false  # disable gitignore support
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// A single diff line can be very long in generated files; the default scanner
// buffer would silently truncate the commit rather than scan it.
const maxDiffLineBytes = 4 * 1024 * 1024

// Version information
const version = "2.2.4"

// Finding represents a detected secret or potential secret
type Finding struct {
	File        string            `json:"file"`
	Line        int               `json:"line"`
	Commit      string            `json:"commit,omitempty"`
	Pattern     string            `json:"pattern"`
	Excerpt     string            `json:"excerpt"`
	RawValue    string            `json:"-"` // Not exported to JSON
	Confidence  float64           `json:"confidence"`
	Verified    bool              `json:"verified"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Hash        string            `json:"hash"`        // Identity of the secret, for deduplication
	Occurrences int               `json:"occurrences"` // Times this same secret was seen
}

// Config holds scanner configuration
type Config struct {
	Rules             map[string]*Rule
	SkipDirs          []string
	SkipFiles         []string
	AllowPatterns     []*regexp.Regexp
	EntropyThreshold  float64
	MinSecretLength   int
	MaxSecretLength   int
	Verbose           bool
	RespectGitignore  bool
	GitignorePatterns []GitignorePattern
	ExcludeGlobs      []string
}

// GitignorePattern represents a pattern from .gitignore with its base directory
type GitignorePattern struct {
	Pattern   string
	Negation  bool
	Directory bool
	BaseDir   string
}

// Rule represents a detection rule
type Rule struct {
	Name        string
	Pattern     *regexp.Regexp
	Keywords    []string
	Description string
	Confidence  float64
	Enabled     bool
}

// Stats tracks scanning statistics
type Stats struct {
	FilesScanned   int
	CommitsScanned int
	FindingsTotal  int
	FindingsUnique int
	StartTime      time.Time
	EndTime        time.Time
	mu             sync.Mutex
}

func (s *Stats) incrementFiles() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.FilesScanned++
}

func (s *Stats) incrementCommits() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CommitsScanned++
}

func (s *Stats) incrementFindings(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.FindingsTotal += count
}

// Enhanced detection patterns with lower false positive rates
var defaultRegexps = map[string]string{
	"aws_access_key":    `AKIA[0-9A-Z]{16}`,
	"aws_secret_key":    `(?i)aws(.{0,20})?(?-i)['\"][0-9a-zA-Z/+]{40}['\"]`,
	"rsa_private":       `-----BEGIN(?: RSA)? PRIVATE KEY-----`,
	"stripe_sk":         `sk_live_[0-9a-zA-Z]{24,}`,
	"stripe_restricted": `rk_live_[0-9a-zA-Z]{24,}`,
	"supabase_jwt":      `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9\.[A-Za-z0-9_\-]+\.[A-Za-z0-9_\-]+`,
	"github_pat":        `ghp_[0-9a-zA-Z]{36}`,
	"github_oauth":      `gho_[0-9a-zA-Z]{36}`,
	"github_app":        `(ghu|ghs)_[0-9a-zA-Z]{36}`,
	"slack_token":       `xox[baprs]-([0-9a-zA-Z]{10,48})`,
	"slack_webhook":     `https://hooks\.slack\.com/services/T[a-zA-Z0-9_]+/B[a-zA-Z0-9_]+/[a-zA-Z0-9_]+`,
	"google_api":        `AIza[0-9A-Za-z_\-]{35}`,
	// A Heroku API key is a bare UUID, so the UUID shape alone carries no
	// signal -- it matched every seeded uuid in every fixture and migration.
	// Require the token to actually be named as a Heroku credential.
	"heroku_api":       `(?i)heroku[a-z0-9_\-]{0,20}['"\s]*[=:>]['"\s]*[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`,
	"mailgun_api":      `key-[0-9a-zA-Z]{32}`,
	"paypal_braintree": `access_token\$production\$[0-9a-z]{16}\$[0-9a-f]{32}`,
	"picatic_api":      `sk_live_[0-9a-z]{32}`,
	"sendgrid_api":     `SG\.[0-9A-Za-z\-_]{22}\.[0-9A-Za-z\-_]{43}`,
	"twilio_api":       `SK[0-9a-fA-F]{32}`,
	"generic_api_key":  `(?i)(?:key|api[_-]?key|apikey)[\s]*[=:>][\s]*['\"]([a-zA-Z0-9_\-]{20,})['\""]`,
	"generic_secret":   `(?i)(?:secret|password|passwd|pwd)[\s]*[=:>][\s]*['\"]([a-zA-Z0-9_\-!@#$%^&*]{8,})['\""]`,
	"db_connection":    `(?i)(postgres|mysql|mongodb|redis)://[^\s'"]+:[^\s'"]+@[^\s'"]+`,
}

// Patterns that should be allowed (common false positives)
var defaultAllowPatterns = []string{
	`^[A-Z_]+$`,                         // All caps constants
	`^[a-z_]+$`,                         // All lowercase
	`(?i)^(true|false|null|undefined)$`, // Boolean/null values
	`^[\d.]+$`,                          // Pure numbers
	`^https?://`,                        // URLs without credentials
	`^[A-Za-z]+\.[A-Za-z]+`,             // Class/module names
	`(?i)^(test|example|sample|demo|placeholder|your[_-].*|my[_-].*)`, // Test values
	`^[*]+$`, // Masked secrets

	// Documentation and fixture credentials advertise themselves, but the
	// marker is usually a suffix or infix rather than a prefix -- an anchored
	// check never saw them. AWS's own canonical example key,
	// AKIAIOSFODNN7EXAMPLE, is the case that made secscan flag its own docs
	// and test fixtures at confidence 0.9.
	`(?i)example`,
	`(?i)(dummy|placeholder|redacted|changeme|notreal|fake)`,
	`(?i)x{8,}`,          // XXXXXXXX-style stand-ins
	`(?i)^(foo|bar|baz)`, // Canonical filler

	// Fill-me-in placeholders from READMEs and .env templates, e.g.
	// JWT_SECRET="your-secret-here". These stay anchored on purpose: an
	// unanchored allow pattern is a silent false negative, because a random
	// 39-character credential can happily contain the substring "my-" or
	// "the_". In a secret scanner an over-broad allowlist is far more dangerous
	// than an over-broad rule -- a false positive is merely annoying, a
	// suppressed real key is invisible.
	`(?i)[_\-]here$`,
	`(?i)^(replace|insert|add)[_\-]?(me|this|your)`,

	// Connection strings whose password is a well-known default are describing
	// a local compose service, not a credential worth reporting.
	`(?i)://[^:/@\s]*:(postgres|mysql|root|password|passwd|admin|secret|redis|mongo|dev|test|guest)@`,
}

// Enhanced skip directories with more comprehensive list
var defaultSkipDirs = []string{
	"node_modules", ".git", "dist", "build", ".next", "venv", "target",
	"__pycache__", ".venv", "env", ".env", "vendor", "coverage",
	".pytest_cache", ".mypy_cache", ".tox", "bin", "obj", ".gradle",
	".idea", ".vscode", ".terraform", "*.egg-info", ".nuxt",
}

// ignoreFileName is a per-repository list of paths that are deliberately
// credential-shaped: fixture files, documentation with sample keys, and so on.
// One glob per line, # for comments, matched against the basename or the
// repo-relative path.
//
// A per-line `secscan:ignore` directive cannot help once a line is committed,
// because history still holds the version without it, and rewriting published
// history to silence a scanner is a bad trade. A repo-level ignore file covers
// the working tree and history alike, which is why history findings carry the
// file they came from.
const ignoreFileName = ".secscanignore"

// Generated files that are build output rather than source.
var defaultSkipArtifacts = []string{
	"search_index.json",                          // MkDocs search index
	"mvnw", "mvnw.cmd", "gradlew", "gradlew.bat", // Build wrappers
}

// File extensions that should be skipped
var defaultSkipExtensions = []string{
	".jpg", ".jpeg", ".png", ".gif", ".ico", ".svg", ".webp",
	".mp4", ".avi", ".mov", ".mp3", ".wav", ".pdf", ".zip",
	".tar", ".gz", ".bz2", ".7z", ".rar", ".exe", ".dll",
	".so", ".dylib", ".bin", ".db", ".sqlite", ".lock",
	".min.js", ".min.css", ".map", ".woff", ".woff2", ".ttf", ".eot",
}

// loadIgnoreFile reads .secscanignore from the scan root. A missing file is
// not an error; most repositories will not have one.
func loadIgnoreFile(root string) []string {
	data, err := os.ReadFile(filepath.Join(root, ignoreFileName))
	if err != nil {
		return nil
	}

	var globs []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		globs = append(globs, line)
	}
	return globs
}

// isExcluded reports whether path matches a .secscanignore glob. It is checked
// against both the basename and the path itself, so `main_test.go` covers the
// file wherever it sits and `docs/*` covers a subtree.
func isExcluded(path string, globs []string) bool {
	if len(globs) == 0 {
		return false
	}

	clean := filepath.ToSlash(path)
	base := filepath.Base(clean)

	for _, glob := range globs {
		if ok, _ := filepath.Match(glob, base); ok {
			return true
		}
		if ok, _ := filepath.Match(glob, clean); ok {
			return true
		}
		// A directory prefix such as "testdata/" or "docs/".
		if strings.HasSuffix(glob, "/") && strings.Contains(clean+"/", glob) {
			return true
		}
	}
	return false
}

func shouldSkipDir(d string) bool {
	base := filepath.Base(d)
	for _, s := range defaultSkipDirs {
		if base == s || strings.HasPrefix(base, s) {
			return true
		}
	}
	return false
}

func shouldSkipFile(name string) bool {
	base := filepath.Base(name)

	// A committed .env is exactly what we are looking for, so it is scanned
	// despite being a hidden file. Its templates are the opposite: .env.example
	// and friends exist to hold placeholders, and reporting them is noise by
	// construction.
	if strings.HasPrefix(base, ".env.") || base == ".env.template" {
		return true
	}

	// Skip hidden files except specific configs
	if strings.HasPrefix(base, ".") && base != ".env" {
		return true
	}

	// Compare against the whole name, not filepath.Ext: Ext("app.min.js")
	// returns ".js", so multi-part suffixes like .min.js and .min.css never
	// matched and minified bundles were scanned as ordinary source -- a large
	// share of the high-entropy false positives came from exactly those files.
	lower := strings.ToLower(base)
	for _, skipExt := range defaultSkipExtensions {
		if strings.HasSuffix(lower, skipExt) {
			return true
		}
	}

	// Skip all types of lock files comprehensively
	lockFilePatterns := []string{
		".lock",             // Gemfile.lock, Pipfile.lock, etc.
		"-lock.json",        // package-lock.json
		"-lock.yaml",        // pnpm-lock.yaml
		"lock.json",         // composer.lock.json
		"lock.yaml",         // yarn.lock (though it's typically yarn.lock)
		"yarn.lock",         // yarn lock file
		"pnpm-lock.yaml",    // pnpm lock file
		"package-lock.json", // npm lock file
		"composer.lock",     // PHP composer lock
		"Gemfile.lock",      // Ruby Gemfile lock
		"Pipfile.lock",      // Python Pipfile lock
		"poetry.lock",       // Python poetry lock
		"Cargo.lock",        // Rust cargo lock
		"mix.lock",          // Elixir mix lock
		"pubspec.lock",      // Dart/Flutter pubspec lock
		"go.sum",            // Go dependencies checksum
		"Podfile.lock",      // iOS CocoaPods lock
		"Cartfile.resolved", // iOS Carthage lock
	}

	for _, pattern := range lockFilePatterns {
		if strings.HasSuffix(base, pattern) || base == pattern {
			return true
		}
	}

	// Generated build artifacts. These are compiled from sources that are
	// themselves scanned, so reading them again adds no coverage -- only noise.
	// A committed docs site is the worst case: `git rev-list --all` reaches the
	// gh-pages branch, and documentation about security is saturated with the
	// words "key", "token" and "secret", so every long string on every rendered
	// page looks like a credential in context.
	for _, artifact := range defaultSkipArtifacts {
		if base == artifact {
			return true
		}
	}

	return false
}

// parseGitignoreLine parses a single line from .gitignore
func parseGitignoreLine(line, baseDir string) *GitignorePattern {
	line = strings.TrimSpace(line)

	// Skip empty lines and comments
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}

	pattern := GitignorePattern{
		BaseDir: baseDir,
	}

	// Check for negation
	if strings.HasPrefix(line, "!") {
		pattern.Negation = true
		line = strings.TrimPrefix(line, "!")
	}

	// Check if pattern is for directories only
	if strings.HasSuffix(line, "/") {
		pattern.Directory = true
		line = strings.TrimSuffix(line, "/")
	}

	pattern.Pattern = line
	return &pattern
}

// loadGitignore loads patterns from a .gitignore file
func loadGitignore(path string) ([]GitignorePattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	baseDir := filepath.Dir(path)
	var patterns []GitignorePattern

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if pattern := parseGitignoreLine(scanner.Text(), baseDir); pattern != nil {
			patterns = append(patterns, *pattern)
		}
	}

	return patterns, scanner.Err()
}

// collectGitignorePatterns finds and loads all .gitignore files in the repository
func collectGitignorePatterns(root string) []GitignorePattern {
	var allPatterns []GitignorePattern

	// Walk the directory tree to find all .gitignore files
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Skip .git directory
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		// Check if this is a .gitignore file
		if !d.IsDir() && d.Name() == ".gitignore" {
			patterns, err := loadGitignore(path)
			if err == nil {
				allPatterns = append(allPatterns, patterns...)
			}
		}

		return nil
	})

	return allPatterns
}

// matchGitignorePattern checks if a path matches a gitignore pattern
func matchGitignorePattern(path string, pattern GitignorePattern) bool {
	// Convert absolute path to relative path from pattern's base directory
	relPath, err := filepath.Rel(pattern.BaseDir, path)
	if err != nil {
		return false
	}

	// If the path is outside the base directory, it doesn't match
	if strings.HasPrefix(relPath, "..") {
		return false
	}

	patternStr := pattern.Pattern

	// Handle different pattern types
	if strings.HasPrefix(patternStr, "/") {
		// Anchored to base directory
		patternStr = strings.TrimPrefix(patternStr, "/")
		return matchPattern(patternStr, relPath)
	} else if strings.Contains(patternStr, "/") {
		// Contains slash - match anywhere in path
		return matchPattern(patternStr, relPath)
	} else {
		// No slash - match basename anywhere in tree
		parts := strings.Split(relPath, string(filepath.Separator))
		for _, part := range parts {
			if matchPattern(patternStr, part) {
				return true
			}
		}
		// Also try full relative path
		return matchPattern(patternStr, relPath)
	}
}

// matchPattern performs glob-style pattern matching
func matchPattern(pattern, name string) bool {
	// Handle ** for matching any number of directories
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := strings.TrimSuffix(parts[0], "/")
			suffix := strings.TrimPrefix(parts[1], "/")

			if prefix != "" && !strings.HasPrefix(name, prefix) {
				return false
			}
			if suffix != "" && !strings.HasSuffix(name, suffix) {
				return false
			}
			return true
		}
	}

	// Simple glob matching
	matched, _ := filepath.Match(pattern, name)
	if matched {
		return true
	}

	// Try matching with the full path for patterns with directory separators
	matched, _ = filepath.Match(pattern, filepath.Base(name))
	return matched
}

// isGitignored checks if a path should be ignored based on gitignore patterns
func isGitignored(path string, patterns []GitignorePattern, isDir bool) bool {
	ignored := false

	// Process patterns in order (later patterns override earlier ones)
	for _, pattern := range patterns {
		// Skip directory-only patterns for files
		if pattern.Directory && !isDir {
			continue
		}

		if matchGitignorePattern(path, pattern) {
			if pattern.Negation {
				ignored = false // Negation pattern - don't ignore
			} else {
				ignored = true // Normal pattern - ignore
			}
		}
	}

	return ignored
}

func looksLikeTextFile(name string) bool {
	// Skip if should skip file
	if shouldSkipFile(name) {
		return false
	}

	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".go", ".js", ".ts", ".tsx", ".jsx", ".java", ".py", ".rb", ".php",
		".json", ".yaml", ".yml", ".env", ".cfg", ".toml", ".md", ".txt",
		".sh", ".bash", ".zsh", ".ps1", ".sql", ".xml", ".html", ".css",
		".c", ".cpp", ".h", ".hpp", ".cs", ".rs", ".kt", ".swift", ".scala",
		".clj", ".ex", ".exs", ".erl", ".hrl", ".vim", ".lua", ".pl", ".r",
		".Dockerfile", ".tf", ".hcl", ".proto", ".graphql", ".vue", ".svelte":
		return true
	default:
		// Check if file has no extension (might be script)
		if ext == "" {
			return true
		}
		return false
	}
}

func compileRules(rules map[string]string) (map[string]*Rule, error) {
	out := make(map[string]*Rule, len(rules))
	for k, v := range rules {
		r, err := regexp.Compile(v)
		if err != nil {
			return nil, fmt.Errorf("failed to compile %s: %w", k, err)
		}
		confidence := 0.9 // Default high confidence for regex patterns
		if strings.HasPrefix(k, "generic_") {
			confidence = 0.7 // Lower confidence for generic patterns
		}
		out[k] = &Rule{
			Name:        k,
			Pattern:     r,
			Description: k,
			Confidence:  confidence,
			Enabled:     true,
		}
	}
	return out, nil
}

func compileAllowPatterns(patterns []string) ([]*regexp.Regexp, error) {
	var out []*regexp.Regexp
	for _, p := range patterns {
		r, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("failed to compile allow pattern %s: %w", p, err)
		}
		out = append(out, r)
	}
	return out, nil
}

// isAllowed checks if a value matches any allow pattern
func isAllowed(value string, allowPatterns []*regexp.Regexp) bool {
	for _, pattern := range allowPatterns {
		if pattern.MatchString(value) {
			return true
		}
	}
	return false
}

// generateHash identifies a finding by the secret itself, not by where it was
// seen. Keying on location (file, or commit SHA) meant the same leaked
// credential produced a distinct hash in every commit that touched it, so
// deduplication across commits could never fire.
func generateHash(pattern, value string) string {
	h := sha256.New()
	h.Write([]byte(pattern + "\x00" + value))
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

// atoiOr parses a base-10 int, falling back to def.
func atoiOr(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func maskSecret(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	pref := s[:4]
	suf := s[len(s)-4:]
	return pref + strings.Repeat("*", len(s)-8) + suf
}

func stringExcerpt(line string, a, b int) string {
	if a < 0 || b > len(line) || a >= b {
		return strings.TrimSpace(line)
	}
	start := a - 20
	if start < 0 {
		start = 0
	}
	end := b + 20
	if end > len(line) {
		end = len(line)
	}
	return strings.TrimSpace(line[start:end])
}

// Secrets live in a fairly narrow length band. Below minSecretLen there is not
// enough material to judge; above maxSecretLen we are almost always looking at
// an embedded asset, a bundled blob, or a base64 payload rather than a
// credential -- and those were the dominant source of false positives.
const (
	minSecretLen = 20
	maxSecretLen = 100

	// defaultEntropy sits just above log2(16) = 4.0, the ceiling for hex, so
	// git SHAs and integrity hashes fall out by construction while base64
	// credentials (which reach ~4.7-5.0 at these lengths) clear it. It must
	// stay below log2(minSecretLen) = 4.32 or the shortest detectable secret
	// becomes mathematically undetectable; TestEntropyThresholdIsReachable
	// enforces that.
	defaultEntropy = 4.1
)

// isHighEntropy reports whether s looks like a randomly generated credential.
//
// Shannon entropy over a single token's own character distribution is bounded
// above by log2(len(s)): a 20-character token cannot score above 4.32 no
// matter how random it is. The old default threshold of 5.0 therefore required
// a token of at least 32 characters with almost no repeated characters, which
// no real credential satisfies -- a live AWS key id scores 3.68, a GitHub PAT
// 4.77, a Stripe key 4.75. The detector could only ever fire on long
// high-alphabet blobs, i.e. exactly the things that are not secrets.
//
// The threshold now sits just above the ceiling for hex (log2(16) = 4.0), so
// "more random than hex" is the bar, and the length band excludes blobs.
func isHighEntropy(s string, threshold float64) bool {
	length := utf8.RuneCountInString(s)
	if length < minSecretLen || length > maxSecretLen {
		return false
	}

	// Must have good character diversity
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSymbol := false

	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			hasLower = true
		} else if r >= 'A' && r <= 'Z' {
			hasUpper = true
		} else if r >= '0' && r <= '9' {
			hasDigit = true
		} else {
			hasSymbol = true
		}
	}

	charClassCount := 0
	for _, v := range []bool{hasLower, hasUpper, hasDigit, hasSymbol} {
		if v {
			charClassCount++
		}
	}

	// Require at least 3 character classes
	if charClassCount < 3 {
		return false
	}

	return shannonEntropy(s) > threshold
}

func shannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	counts := make(map[rune]int)
	for _, r := range s {
		counts[r]++
	}
	e := 0.0
	L := float64(len(s))
	for _, c := range counts {
		p := float64(c) / L
		e += -p * math.Log2(p)
	}
	return e
}

func walkFiles(root string, config *Config, action func(path string) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// ignore walk errors for robustness
			return nil
		}
		if d.IsDir() {
			// Check gitignore first if enabled
			if config.RespectGitignore && isGitignored(path, config.GitignorePatterns, true) {
				if config.Verbose {
					fmt.Printf("Skipping gitignored directory: %s\n", path)
				}
				return filepath.SkipDir
			}

			// Then check default skip dirs
			if shouldSkipDir(path) && path != root {
				return filepath.SkipDir
			}
			return nil
		}

		// Check gitignore for files if enabled
		if config.RespectGitignore && isGitignored(path, config.GitignorePatterns, false) {
			if config.Verbose {
				fmt.Printf("Skipping gitignored file: %s\n", path)
			}
			return nil
		}

		if !looksLikeTextFile(path) {
			return nil
		}
		if isExcluded(path, config.ExcludeGlobs) {
			if config.Verbose {
				fmt.Printf("Skipping excluded file: %s\n", path)
			}
			return nil
		}
		return action(path)
	})
}

// entropyTokens extracts candidate secrets: maximal runs in the base64url
// alphabet, which is what credentials are actually drawn from.
//
// This used to be `\S{20,}` -- any run of non-whitespace. That swallowed
// whole URLs, HTML attributes and code fragments as single "tokens", so
// `href="../user-guide/examples/#share"` was handed to the entropy check as if
// it were a candidate credential. Restricting the alphabet drops separators,
// quotes and punctuation, which splits structured text back into short words
// while leaving real tokens intact.
var entropyTokens = regexp.MustCompile(`[A-Za-z0-9_\-]{20,}`)

// secretContext matches lines that name something as a credential.
//
// Entropy on its own is not a detector at any threshold. Set it high enough to
// reject ordinary source and it rejects real credentials too (Shannon entropy
// over a token's own characters is bounded by log2(len), so a 20-char secret
// tops out at 4.32); set it low enough to catch them and every minified
// fragment, URL and base64 chunk in the tree fires. Measured on a 37-repo
// corpus: a 5.0 threshold found 0 real secrets, and 4.1 produced 15,387
// findings.
//
// What separates a credential from a random-looking token is not its entropy
// but that somebody named it. Consult entropy only where the line says the
// value is a secret, and it becomes a useful last-resort detector for
// credential formats that have no dedicated rule.
var secretContext = regexp.MustCompile(`(?i)(secret|token|password|passwd|` +
	`\bpwd\b|api[_\-.]?key|apikey|access[_\-.]?key|private[_\-.]?key|` +
	`credential|auth|bearer|client[_\-.]?secret|signing[_\-.]?key|` +
	`encryption[_\-.]?key|\bsalt\b|passphrase)`)

// ignoreDirective marks a line that is deliberately credential-shaped.
//
// Some values cannot be allowlisted by their content without weakening the
// allowlist for everyone: a scanner's own test suite has to contain realistic,
// non-allowlisted credentials in order to prove that real ones are not
// suppressed. Widening the allow patterns to cover them would create silent
// false negatives, which is the one failure mode worse than noise. An explicit
// per-line opt-out is the honest way to say "I know what this is".
//
//	{"github pat", "ghp_...", true}, // secscan:ignore
var ignoreDirective = regexp.MustCompile(`(?i)secscan:\s*ignore`)

// scanLine applies every rule plus entropy detection to a single line of
// content. Both the working-tree scanner and the git-history scanner go
// through here, so a secret is judged the same way wherever it is found.
func scanLine(line, file string, lineNo int, rules map[string]*Rule, config *Config) []Finding {
	trim := strings.TrimSpace(line)
	if trim == "" {
		return nil
	}

	if ignoreDirective.MatchString(line) {
		return nil
	}

	// Skip comments (basic detection)
	if strings.HasPrefix(trim, "//") || strings.HasPrefix(trim, "#") ||
		strings.HasPrefix(trim, "/*") || strings.HasPrefix(trim, "*") {
		return nil
	}

	var findings []Finding

	for name, rule := range rules {
		if !rule.Enabled {
			continue
		}

		loc := rule.Pattern.FindStringSubmatchIndex(line)
		if loc == nil {
			continue
		}

		// Test the allowlist against the captured secret, not the whole match.
		// Rules like generic_secret match `password="testpass123"` in full, so
		// checking the match meant anchored allow patterns such as ^test or
		// ^your[_-] could never fire -- every one of them was dead code, and
		// obvious placeholders were reported as credentials.
		rawValue := line[loc[0]:loc[1]]
		value := rawValue
		if len(loc) >= 4 && loc[2] >= 0 {
			value = line[loc[2]:loc[3]]
		}

		if isAllowed(value, config.AllowPatterns) || isAllowed(rawValue, config.AllowPatterns) {
			continue
		}

		findings = append(findings, Finding{
			File:       file,
			Line:       lineNo,
			Pattern:    name,
			Excerpt:    maskSecret(stringExcerpt(line, loc[0], loc[1])),
			RawValue:   rawValue,
			Confidence: rule.Confidence,
			Hash:       generateHash(name, rawValue),
		})
	}

	if config.EntropyThreshold > 0 && secretContext.MatchString(line) {
		for _, tok := range entropyTokens.FindAllString(line, -1) {
			// The token that matched the context keyword is the name, not the
			// value; don't report the identifier as its own secret.
			if secretContext.MatchString(tok) {
				continue
			}
			if isAllowed(tok, config.AllowPatterns) {
				continue
			}
			if isHighEntropy(tok, config.EntropyThreshold) {
				findings = append(findings, Finding{
					File:       file,
					Line:       lineNo,
					Pattern:    "high_entropy",
					Excerpt:    maskSecret(tok),
					RawValue:   tok,
					Confidence: 0.6,
					Hash:       generateHash("high_entropy", tok),
				})
			}
		}
	}

	return findings
}

func scanFileForSecrets(path string, rules map[string]*Rule, config *Config) ([]Finding, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var findings []Finding
	r := bufio.NewReader(f)
	lineNo := 0

	for {
		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return findings, err
		}
		lineNo++

		findings = append(findings, scanLine(line, path, lineNo, rules, config)...)

		if err == io.EOF {
			break
		}
	}
	return findings, nil
}

// git helpers
func gitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// isGitRepo reports whether root is inside a git working tree.
func isGitRepo(root string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = root
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

func gitAllCommits(root string) ([]string, error) {
	cmd := exec.Command("git", "rev-list", "--all")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git rev-list failed: %w (%s)", err, out)
	}
	lines := strings.Fields(string(out))
	return lines, nil
}

func gitShowCommitDiff(root, commit string) (string, error) {
	cmd := exec.Command("git", "show", "--pretty=", "--unified=0", commit)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git show %s failed: %w", commit, err)
	}
	return string(out), nil
}

// hunkHeader matches "@@ -12,3 +14,2 @@", capturing the old and new start lines.
var hunkHeader = regexp.MustCompile(`^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

func scanGitHistory(root string, rules map[string]*Rule, config *Config, stats *Stats) ([]Finding, error) {
	if !gitAvailable() {
		return nil, errors.New("git not available in PATH")
	}
	if !isGitRepo(root) {
		return nil, fmt.Errorf("%s is not a git repository", root)
	}
	commits, err := gitAllCommits(root)
	if err != nil {
		return nil, err
	}

	var results []Finding
	for _, c := range commits {
		diff, err := gitShowCommitDiff(root, c)
		if err != nil {
			// skip commits that fail
			continue
		}

		stats.incrementCommits()

		s := bufio.NewScanner(strings.NewReader(diff))
		s.Buffer(make([]byte, 0, 64*1024), maxDiffLineBytes)
		currentFile := ""
		skipCurrentFile := false
		oldLine, newLine := 0, 0

		for s.Scan() {
			line := s.Text()

			// Git diff format: "diff --git a/path/to/file b/path/to/file"
			if strings.HasPrefix(line, "diff --git") {
				parts := strings.Fields(line)
				if len(parts) >= 4 {
					// parts[3] is b/filename, the new version of the path.
					currentFile = strings.TrimPrefix(parts[3], "b/")
					skipCurrentFile = shouldSkipFile(currentFile) ||
						isExcluded(currentFile, config.ExcludeGlobs)

					if config.Verbose && skipCurrentFile {
						fmt.Printf("Skipping file in git history: %s (commit: %s)\n", currentFile, c[:8])
					}
				}
				continue
			}

			// Hunk headers carry the line numbers a finding should point at.
			if m := hunkHeader.FindStringSubmatch(line); m != nil {
				oldLine = atoiOr(m[1], 0)
				newLine = atoiOr(m[2], 0)
				continue
			}

			// The ---/+++ path headers are not content.
			if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
				continue
			}

			// Only added or removed lines carry changed content. Track the line
			// number in whichever version of the file the line belongs to.
			var lineNo int
			switch {
			case strings.HasPrefix(line, "+"):
				lineNo = newLine
				newLine++
			case strings.HasPrefix(line, "-"):
				lineNo = oldLine
				oldLine++
			default:
				oldLine++
				newLine++
				continue
			}

			if skipCurrentFile {
				continue
			}

			// Strip the diff marker so the content scans identically to a
			// working-tree line -- including comment detection.
			for _, f := range scanLine(line[1:], currentFile, lineNo, rules, config) {
				f.Commit = c
				results = append(results, f)
			}
		}
	}
	return results, nil
}

// isJWTSegment reports whether s is one base64url segment of a JWT -- that is,
// whether it decodes to a JSON object. The header and payload of every JWT do;
// random credentials do not.
func isJWTSegment(s string) bool {
	if !strings.HasPrefix(s, "eyJ") {
		return false
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "="))
	if err != nil {
		return false
	}
	var obj map[string]interface{}
	return json.Unmarshal(raw, &obj) == nil
}

// jwtClaims decodes the claims payload of a compact-serialized JWT. It does
// not verify the signature -- we only need to read what the token asserts
// about itself.
func jwtClaims(token string) (map[string]interface{}, bool) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(parts[1], "="))
	if err != nil {
		return nil, false
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(raw, &claims); err != nil {
		return nil, false
	}
	return claims, true
}

// refineFindings applies structural checks that a regular expression cannot
// express: it drops matches that are public by design and promotes the ones
// that can be shown to be sensitive. This is what populates Verified, which
// until now was written as false on every finding and never revisited.
func refineFindings(findings []Finding) []Finding {
	out := findings[:0]

	for _, f := range findings {
		// A JWT is three base64 segments joined by dots, so the entropy
		// tokenizer splits one token into three and reports each separately.
		// A segment that decodes to a JSON object is part of a JWT, and the
		// dedicated rule already owns the whole token.
		if f.Pattern == "high_entropy" && isJWTSegment(f.RawValue) {
			continue
		}

		// Any rule can surface a JWT -- the dedicated supabase_jwt rule, or the
		// entropy detector picking the same token out of the same line. Judge
		// the token by what it is rather than by which rule happened to see it.
		if strings.HasPrefix(f.RawValue, "eyJ") {
			claims, ok := jwtClaims(f.RawValue)
			if !ok {
				if f.Pattern == "supabase_jwt" {
					continue // Matched the rule but isn't actually a JWT.
				}
				out = append(out, f)
				continue
			}

			role, _ := claims["role"].(string)
			switch role {
			case "anon":
				// Supabase anon keys are designed to be shipped to browsers and
				// are protected by row-level security, not by secrecy. Reporting
				// one is always a false positive.
				continue
			case "service_role":
				// This one bypasses row-level security entirely.
				f.Verified = true
				f.Confidence = 0.99
				if f.Metadata == nil {
					f.Metadata = map[string]string{}
				}
				f.Metadata["role"] = role
				f.Metadata["reason"] = "service_role bypasses row-level security"
			}

			if iss, _ := claims["iss"].(string); iss == "supabase-demo" {
				// The fixed key from `supabase init`, published in Supabase's docs.
				continue
			}
		}

		out = append(out, f)
	}

	return out
}

// deduplicateFindings collapses findings that refer to the same secret. The
// first occurrence wins and carries a count of everywhere else it was seen.
func deduplicateFindings(findings []Finding) []Finding {
	index := make(map[string]int)
	var unique []Finding

	for _, f := range findings {
		if i, ok := index[f.Hash]; ok {
			unique[i].Occurrences++
			// Prefer to report a secret at a location someone can still act on.
			if unique[i].Commit != "" && f.Commit == "" {
				occ := unique[i].Occurrences
				unique[i] = f
				unique[i].Occurrences = occ
			}
			continue
		}
		f.Occurrences = 1
		index[f.Hash] = len(unique)
		unique = append(unique, f)
	}

	return unique
}

func printFindings(findings []Finding, verbose bool) {
	if len(findings) == 0 {
		fmt.Println("✅ No secrets found")
		return
	}

	// Sort by file, then line
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].File == findings[j].File {
			return findings[i].Line < findings[j].Line
		}
		return findings[i].File < findings[j].File
	})

	// Group by severity
	critical := 0
	high := 0
	medium := 0
	low := 0

	for _, f := range findings {
		switch {
		case f.Confidence >= 0.9:
			critical++
		case f.Confidence >= 0.8:
			high++
		case f.Confidence >= 0.6:
			medium++
		default:
			low++
		}
	}

	fmt.Println("\n🔍 Secret Scan Results")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("Total findings: %d\n", len(findings))
	fmt.Printf("  Critical (≥0.9): %d\n", critical)
	fmt.Printf("  High (≥0.8):     %d\n", high)
	fmt.Printf("  Medium (≥0.6):   %d\n", medium)
	fmt.Printf("  Low (<0.6):      %d\n", low)
	fmt.Println("=" + strings.Repeat("=", 50))

	if !verbose && len(findings) > 100 {
		fmt.Printf("\nShowing first 100 findings (use -verbose to see all)\n\n")
		findings = findings[:100]
	}

	for _, f := range findings {
		// Color coding based on confidence
		var prefix string
		switch {
		case f.Confidence >= 0.9:
			prefix = "🔴 [CRITICAL]"
		case f.Confidence >= 0.8:
			prefix = "🟠 [HIGH]"
		case f.Confidence >= 0.6:
			prefix = "🟡 [MEDIUM]"
		default:
			prefix = "⚪ [LOW]"
		}

		fmt.Printf("%s [%s] %s:%d", prefix, strings.ToUpper(f.Pattern), f.File, f.Line)
		if f.Commit != "" {
			fmt.Printf(" (commit %s)", f.Commit[:8])
		}
		fmt.Printf("\n  → %s (confidence: %.2f)\n", f.Excerpt, f.Confidence)

		if verbose && f.Verified {
			fmt.Println("  ✓ Verified")
		}
		fmt.Println()
	}
}

func loadRulesFromFile(path string) (map[string]string, error) {
	// very small TOML-like parser: key = "regex" per line
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, l := range strings.Split(string(b), "\n") {
		line := strings.TrimSpace(l)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		v = strings.Trim(v, " \"")
		m[k] = v
	}
	return m, nil
}

func main() {
	// Command line flags
	root := flag.String("root", ".", "project root to scan")
	history := flag.Bool("history", true, "scan git history (slower)")
	jsonOut := flag.String("json", "", "path to write JSON report (optional)")
	quiet := flag.Bool("quiet", false, "suppress human output (useful for CI)")
	verbose := flag.Bool("verbose", false, "show detailed output with all findings")
	configFile := flag.String("config", "", "path to custom config file (optional)")
	entropyThreshold := flag.Float64("entropy", defaultEntropy, "entropy threshold for detection")
	noEntropy := flag.Bool("no-entropy", false, "disable entropy-based detection")
	showVersion := flag.Bool("version", false, "show version information")
	respectGitignore := flag.Bool("respect-gitignore", true, "respect .gitignore files when scanning (default: true)")
	minConfidence := flag.Float64("min-confidence", 0, "only report findings at or above this confidence (0-1)")

	flag.Parse()

	historyFailed := false

	if *showVersion {
		fmt.Printf("secscan version %s\n", version)
		fmt.Println("Enhanced secret scanner for source code")
		fmt.Println("https://github.com/Zayan-Mohamed/secscan")
		os.Exit(0)
	}

	// Initialize stats
	stats := &Stats{
		StartTime: time.Now(),
	}

	// Load rules
	var rulesMap map[string]string
	if *configFile != "" {
		loaded, err := loadRulesFromFile(*configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load config file: %v\n", err)
			fmt.Fprintf(os.Stderr, "Using default rules...\n")
			rulesMap = defaultRegexps
		} else {
			rulesMap = loaded
		}
	} else {
		rulesMap = defaultRegexps
	}

	// Compile rules
	compiled, err := compileRules(rulesMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to compile rules: %v\n", err)
		os.Exit(2)
	}

	// Compile allow patterns
	allowPatterns, err := compileAllowPatterns(defaultAllowPatterns)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to compile allow patterns: %v\n", err)
		os.Exit(2)
	}

	// Load gitignore patterns if enabled
	var gitignorePatterns []GitignorePattern
	if *respectGitignore {
		gitignorePatterns = collectGitignorePatterns(*root)
		if !*quiet && len(gitignorePatterns) > 0 {
			fmt.Printf("Loaded %d .gitignore patterns\n", len(gitignorePatterns))
		}
	}

	excludeGlobs := loadIgnoreFile(*root)
	if len(excludeGlobs) > 0 && !*quiet {
		fmt.Printf("Loaded %d %s patterns\n", len(excludeGlobs), ignoreFileName)
	}

	// Create config
	config := &Config{
		Rules:             compiled,
		AllowPatterns:     allowPatterns,
		EntropyThreshold:  *entropyThreshold,
		MinSecretLength:   8,
		MaxSecretLength:   512,
		Verbose:           *verbose,
		RespectGitignore:  *respectGitignore,
		GitignorePatterns: gitignorePatterns,
		ExcludeGlobs:      excludeGlobs,
	}

	if *noEntropy {
		config.EntropyThreshold = 0
	}

	if !*quiet {
		fmt.Printf("SecScan v%s - Enhanced Secret Scanner\n", version)
		fmt.Printf("Scanning: %s\n", *root)
		fmt.Printf("Entropy threshold: %.1f\n", config.EntropyThreshold)
		fmt.Printf("Rules loaded: %d\n", len(compiled))
		if *respectGitignore {
			fmt.Printf("Gitignore: enabled (%d patterns loaded)\n", len(gitignorePatterns))
		} else {
			fmt.Println("Gitignore: disabled")
		}
		if *history {
			fmt.Println("Git history: enabled")
		}
		fmt.Println()
	}

	var allFindings []Finding

	// Scan files
	_ = walkFiles(*root, config, func(path string) error {
		fnds, err := scanFileForSecrets(path, compiled, config)
		if err != nil {
			// ignore read errors on a file
			return nil
		}
		if len(fnds) > 0 {
			allFindings = append(allFindings, fnds...)
			stats.incrementFindings(len(fnds))
		}
		stats.incrementFiles()
		return nil
	})

	// Scan git history
	if *history {
		gh, err := scanGitHistory(*root, compiled, config, stats)
		if err != nil {
			// This used to be swallowed whenever it produced no findings, so a
			// history scan that never ran was indistinguishable from a clean
			// one. It is the whole point of the tool: say so loudly.
			fmt.Fprintf(os.Stderr, "Warning: git history scan failed: %v\n", err)
			historyFailed = true
		} else if len(gh) > 0 {
			allFindings = append(allFindings, gh...)
			stats.incrementFindings(len(gh))
		}
	}

	// Deduplicate findings
	uniqueFindings := refineFindings(allFindings)
	uniqueFindings = deduplicateFindings(uniqueFindings)

	if *minConfidence > 0 {
		kept := uniqueFindings[:0]
		for _, f := range uniqueFindings {
			if f.Confidence >= *minConfidence {
				kept = append(kept, f)
			}
		}
		uniqueFindings = kept
	}

	stats.FindingsUnique = len(uniqueFindings)
	stats.EndTime = time.Now()

	// Write JSON output
	if *jsonOut != "" {
		output := map[string]interface{}{
			"findings": uniqueFindings,
			"stats": map[string]interface{}{
				"files_scanned":    stats.FilesScanned,
				"commits_scanned":  stats.CommitsScanned,
				"findings_total":   stats.FindingsTotal,
				"findings_unique":  stats.FindingsUnique,
				"scan_duration_ms": stats.EndTime.Sub(stats.StartTime).Milliseconds(),
				"history_scanned":  *history && !historyFailed,
			},
			"version": version,
		}
		b, _ := json.MarshalIndent(output, "", "  ")
		_ = os.WriteFile(*jsonOut, b, 0644)

		if !*quiet {
			fmt.Printf("JSON report written to: %s\n\n", *jsonOut)
		}
	}

	// Print human-readable output
	if !*quiet {
		printFindings(uniqueFindings, *verbose)

		fmt.Println("\n Scan Statistics")
		fmt.Println("=" + strings.Repeat("=", 50))
		fmt.Printf("Files scanned:    %d\n", stats.FilesScanned)
		if *history {
			fmt.Printf("Commits scanned:  %d\n", stats.CommitsScanned)
		}
		fmt.Printf("Total findings:   %d\n", stats.FindingsTotal)
		fmt.Printf("Unique findings:  %d\n", stats.FindingsUnique)
		fmt.Printf("Scan duration:    %v\n", stats.EndTime.Sub(stats.StartTime).Round(time.Millisecond))
		fmt.Println("=" + strings.Repeat("=", 50))
	}

	// Exit with error code if secrets found
	if len(uniqueFindings) > 0 {
		os.Exit(1)
	}
	// "No secrets found" is only meaningful if we actually looked. Exiting 0
	// after a failed history scan reports a clean bill of health for a scan
	// that never happened.
	if historyFailed {
		os.Exit(2)
	}
	os.Exit(0)
}
