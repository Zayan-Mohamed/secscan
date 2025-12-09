// secscan - Enhanced Go CLI secret scanner
// Version: 2.1.0
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
//	secscan -root . -entropy 5.5         # adjust entropy threshold
//	secscan -root . -verbose             # show detailed output
//	secscan -root . -respect-gitignore=false  # disable gitignore support
package main

import (
	"bufio"
	"crypto/sha256"
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
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// Finding represents a detected secret or potential secret
type Finding struct {
	File       string            `json:"file"`
	Line       int               `json:"line"`
	Commit     string            `json:"commit,omitempty"`
	Pattern    string            `json:"pattern"`
	Excerpt    string            `json:"excerpt"`
	RawValue   string            `json:"-"` // Not exported to JSON
	Confidence float64           `json:"confidence"`
	Verified   bool              `json:"verified"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Hash       string            `json:"hash"` // For deduplication
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
	"heroku_api":        `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`,
	"mailgun_api":       `key-[0-9a-zA-Z]{32}`,
	"paypal_braintree":  `access_token\$production\$[0-9a-z]{16}\$[0-9a-f]{32}`,
	"picatic_api":       `sk_live_[0-9a-z]{32}`,
	"sendgrid_api":      `SG\.[0-9A-Za-z\-_]{22}\.[0-9A-Za-z\-_]{43}`,
	"twilio_api":        `SK[0-9a-fA-F]{32}`,
	"generic_api_key":   `(?i)(?:key|api[_-]?key|apikey)[\s]*[=:>][\s]*['\"]([a-zA-Z0-9_\-]{20,})['\""]`,
	"generic_secret":    `(?i)(?:secret|password|passwd|pwd)[\s]*[=:>][\s]*['\"]([a-zA-Z0-9_\-!@#$%^&*]{8,})['\""]`,
	"db_connection":     `(?i)(postgres|mysql|mongodb|redis)://[^\s'"]+:[^\s'"]+@[^\s'"]+`,
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
}

// Enhanced skip directories with more comprehensive list
var defaultSkipDirs = []string{
	"node_modules", ".git", "dist", "build", ".next", "venv", "target",
	"__pycache__", ".venv", "env", ".env", "vendor", "coverage",
	".pytest_cache", ".mypy_cache", ".tox", "bin", "obj", ".gradle",
	".idea", ".vscode", ".terraform", "*.egg-info", ".nuxt",
}

// File extensions that should be skipped
var defaultSkipExtensions = []string{
	".jpg", ".jpeg", ".png", ".gif", ".ico", ".svg", ".webp",
	".mp4", ".avi", ".mov", ".mp3", ".wav", ".pdf", ".zip",
	".tar", ".gz", ".bz2", ".7z", ".rar", ".exe", ".dll",
	".so", ".dylib", ".bin", ".db", ".sqlite", ".lock",
	".min.js", ".min.css", ".map", ".woff", ".woff2", ".ttf", ".eot",
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

	// Skip hidden files except specific configs
	if strings.HasPrefix(base, ".") && base != ".env" && base != ".env.example" {
		return true
	}

	ext := strings.ToLower(filepath.Ext(name))
	for _, skipExt := range defaultSkipExtensions {
		if ext == skipExt {
			return true
		}
	}

	// Skip lock files
	if strings.HasSuffix(base, ".lock") || strings.HasSuffix(base, "-lock.json") {
		return true
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

// generateHash creates a unique hash for deduplication
func generateHash(file, pattern, value string) string {
	h := sha256.New()
	h.Write([]byte(file + pattern + value))
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
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

// Improved entropy detection with adjustable threshold
func isHighEntropy(s string, threshold float64) bool {
	length := utf8.RuneCountInString(s)

	// Must be at least 20 characters
	if length < 20 {
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

	// Calculate Shannon entropy
	h := shannonEntropy(s)

	// Check against threshold (default 5.0, increased from 4.0)
	return h > threshold
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
		return action(path)
	})
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
		trim := strings.TrimSpace(line)
		if len(trim) == 0 {
			if err == io.EOF {
				break
			}
			continue
		}

		// Skip comments (basic detection)
		if strings.HasPrefix(trim, "//") || strings.HasPrefix(trim, "#") ||
			strings.HasPrefix(trim, "/*") || strings.HasPrefix(trim, "*") {
			if err == io.EOF {
				break
			}
			continue
		}

		// Check regex rules
		for name, rule := range rules {
			if !rule.Enabled {
				continue
			}

			if loc := rule.Pattern.FindStringIndex(line); loc != nil {
				excerpt := stringExcerpt(line, loc[0], loc[1])
				rawValue := line[loc[0]:loc[1]]

				// Check if allowed
				if isAllowed(rawValue, config.AllowPatterns) {
					continue
				}

				findings = append(findings, Finding{
					File:       path,
					Line:       lineNo,
					Pattern:    name,
					Excerpt:    maskSecret(excerpt),
					RawValue:   rawValue,
					Confidence: rule.Confidence,
					Verified:   false,
					Hash:       generateHash(path, name, rawValue),
				})
			}
		}

		// Check high entropy tokens
		if config.EntropyThreshold > 0 {
			tokens := regexp.MustCompile(`\S{20,}`).FindAllString(line, -1)
			for _, tok := range tokens {
				// Skip if allowed
				if isAllowed(tok, config.AllowPatterns) {
					continue
				}

				if isHighEntropy(tok, config.EntropyThreshold) {
					findings = append(findings, Finding{
						File:       path,
						Line:       lineNo,
						Pattern:    "high_entropy",
						Excerpt:    maskSecret(tok),
						RawValue:   tok,
						Confidence: 0.6,
						Verified:   false,
						Hash:       generateHash(path, "high_entropy", tok),
					})
				}
			}
		}

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

func gitAllCommits() ([]string, error) {
	cmd := exec.Command("git", "rev-list", "--all")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git rev-list failed: %w (%s)", err, out)
	}
	lines := strings.Fields(string(out))
	return lines, nil
}

func gitShowCommitDiff(commit string) (string, error) {
	cmd := exec.Command("git", "show", "--pretty=", "--unified=0", commit)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git show %s failed: %w", commit, err)
	}
	return string(out), nil
}

func scanGitHistory(rules map[string]*Rule, config *Config, stats *Stats) ([]Finding, error) {
	if !gitAvailable() {
		return nil, errors.New("git not available in PATH")
	}
	commits, err := gitAllCommits()
	if err != nil {
		return nil, err
	}

	var results []Finding
	for _, c := range commits {
		diff, err := gitShowCommitDiff(c)
		if err != nil {
			// skip commits that fail
			continue
		}

		stats.incrementCommits()

		s := bufio.NewScanner(strings.NewReader(diff))
		ln := 0
		for s.Scan() {
			ln++
			line := s.Text()

			// Only scan added or removed lines
			if !strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "-") {
				continue
			}

			// Check regex patterns
			for name, rule := range rules {
				if !rule.Enabled {
					continue
				}

				if loc := rule.Pattern.FindStringIndex(line); loc != nil {
					rawValue := line[loc[0]:loc[1]]

					// Check if allowed
					if isAllowed(rawValue, config.AllowPatterns) {
						continue
					}

					results = append(results, Finding{
						File:       "(git-history)",
						Line:       ln,
						Commit:     c,
						Pattern:    name,
						Excerpt:    maskSecret(stringExcerpt(line, loc[0], loc[1])),
						RawValue:   rawValue,
						Confidence: 0.85,
						Verified:   false,
						Hash:       generateHash(c, name, rawValue),
					})
				}
			}

			// Check entropy (with stricter threshold for git history)
			if config.EntropyThreshold > 0 {
				toks := regexp.MustCompile(`\S{20,}`).FindAllString(line, -1)
				for _, t := range toks {
					// Skip if allowed
					if isAllowed(t, config.AllowPatterns) {
						continue
					}

					// Use higher threshold for git history to reduce noise
					if isHighEntropy(t, config.EntropyThreshold+0.5) {
						results = append(results, Finding{
							File:       "(git-history)",
							Line:       ln,
							Commit:     c,
							Pattern:    "high_entropy",
							Excerpt:    maskSecret(t),
							RawValue:   t,
							Confidence: 0.55,
							Verified:   false,
							Hash:       generateHash(c, "high_entropy", t),
						})
					}
				}
			}
		}
	}
	return results, nil
}

// deduplicateFindings removes duplicate findings based on hash
func deduplicateFindings(findings []Finding) []Finding {
	seen := make(map[string]bool)
	var unique []Finding

	for _, f := range findings {
		if !seen[f.Hash] {
			seen[f.Hash] = true
			unique = append(unique, f)
		}
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
	entropyThreshold := flag.Float64("entropy", 5.0, "entropy threshold for detection (default 5.0)")
	noEntropy := flag.Bool("no-entropy", false, "disable entropy-based detection")
	version := flag.Bool("version", false, "show version information")
	respectGitignore := flag.Bool("respect-gitignore", true, "respect .gitignore files when scanning (default: true)")

	flag.Parse()

	if *version {
		fmt.Println("secscan version 2.1.0")
		fmt.Println("Enhanced secret scanner for source code")
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
	}

	if *noEntropy {
		config.EntropyThreshold = 0
	}

	if !*quiet {
		fmt.Println("SecScan v2.1.0 - Enhanced Secret Scanner")
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
		gh, err := scanGitHistory(compiled, config, stats)
		if err == nil && len(gh) > 0 {
			allFindings = append(allFindings, gh...)
			stats.incrementFindings(len(gh))
		} else if err != nil && !*quiet {
			fmt.Fprintf(os.Stderr, "Warning: git history scan failed: %v\n", err)
		}
	}

	// Deduplicate findings
	uniqueFindings := deduplicateFindings(allFindings)
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
			},
			"version": "2.1.0",
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
	os.Exit(0)
}
