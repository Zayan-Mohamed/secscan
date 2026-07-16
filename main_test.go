package main

import (
	"encoding/base64"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

		// filepath.Ext("app.min.js") is ".js", so comparing against a
		// multi-part suffix never matched and minified bundles were scanned as
		// ordinary source -- a major source of high-entropy false positives.
		{"app.min.js", true},
		{"styles.min.css", true},
		{"bundle.js.map", true},
		{"app.js", false},
		{"styles.css", false},
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

// TestIsHighEntropy verifies high entropy detection.
//
// The previous version of this test logged mismatches instead of failing them,
// so it passed unconditionally -- which is how a detector that could not fire
// on any real credential shipped green.
func TestIsHighEntropy(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		threshold float64
		want      bool
	}{
		{"repeated chars", "aaaaaaaaaaaaaaaaaaaaaaaa", defaultEntropy, false},
		{"too short", "test123", defaultEntropy, false},
		{"english prose", "the quick brown fox jumped over", defaultEntropy, false},

		// Hex cannot exceed log2(16) = 4.0, so git SHAs and integrity hashes
		// fall below the threshold by construction rather than by luck.
		{"git sha", "5f2d8a1c9b3e7f04a6d2c8b1e9f3a7d05c2b8e14", defaultEntropy, false},

		// Real credentials must actually fire. Under the old 5.0 default every
		// one of these scored below threshold and was silently missed. The PAT
		// scores 4.77, which is the band real credentials actually occupy.
		//
		// These fixtures deliberately avoid live provider prefixes. A literal
		// sk_live_ token here is blocked by GitHub's push protection even
		// though it is a fixture -- which is the same false positive this
		// release fixes in secscan, where an anchored allowlist missed AWS's
		// own AKIAIOSFODNN7EXAMPLE and the scanner flagged its own tests.
		{"github pat", "ghp_a1B2c3D4e5F6g7H8i9J0k1L2m3N4o5P6q7R8", defaultEntropy, true}, // secscan:ignore
		{"opaque service key", "vault_k3Qm9RtZ2xPwL7bNcF4hJ8sVyD6g", defaultEntropy, true},

		// Long base64 blobs are embedded assets, not secrets. These were the
		// dominant false positive.
		{"embedded blob", strings.Repeat("aGVsbG9Xb3JsZDEyMzQ1Njc4OTBhYmNkZWY", 6), defaultEntropy, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isHighEntropy(tt.input, tt.threshold); got != tt.want {
				t.Errorf("isHighEntropy(%q) = %v, want %v (entropy=%.2f, len=%d)",
					tt.input, got, tt.want, shannonEntropy(tt.input), len(tt.input))
			}
		})
	}
}

// TestEntropyThresholdIsReachable guards the arithmetic that broke the
// detector: Shannon entropy over a token's own characters is bounded by
// log2(len), so a threshold of 5.0 silently required a 32+ character token
// with almost no repeats -- a bar no real credential clears.
func TestEntropyThresholdIsReachable(t *testing.T) {
	ceiling := math.Log2(float64(minSecretLen))
	if defaultEntropy >= ceiling {
		t.Fatalf("default entropy %.2f is unreachable for a %d-char token (max possible %.2f): "+
			"no secret of the minimum length can ever be detected",
			defaultEntropy, minSecretLen, ceiling)
	}
}

// TestHerokuRequiresContext pins that a bare UUID is not a Heroku key. The
// pattern used to be the UUID shape alone, which matched every seeded uuid in
// every fixture and SQL migration.
func TestHerokuRequiresContext(t *testing.T) {
	rules, err := compileRules(defaultRegexps)
	if err != nil {
		t.Fatal(err)
	}
	rule := rules["heroku_api"]

	if rule.Pattern.MatchString(`INSERT INTO users VALUES ('550e8400-e29b-41d4-a716-446655440000', 'john')`) {
		t.Error("a bare UUID in a SQL fixture must not be reported as a Heroku key")
	}
	if !rule.Pattern.MatchString(`HEROKU_API_KEY="550e8400-e29b-41d4-a716-446655440000"`) { // secscan:ignore
		t.Error("a UUID named as a Heroku key should still be reported")
	}
}

// TestExampleCredentialsAllowed pins that documentation and fixture keys are
// not reported. The allowlist was anchored with ^, so it never matched AWS's
// canonical AKIAIOSFODNN7EXAMPLE -- and secscan flagged its own docs and test
// fixtures at confidence 0.9.
func TestExampleCredentialsAllowed(t *testing.T) {
	allow, err := compileAllowPatterns(defaultAllowPatterns)
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range []string{
		"AKIAIOSFODNN7EXAMPLE",
		"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"postgres://postgres:postgres@db:5432/app",
	} {
		if !isAllowed(v, allow) {
			t.Errorf("%q should be allowlisted as a documentation/default credential", v)
		}
	}

	if isAllowed("AKIAI44QH8DHBEXMPLZZ", allow) { // secscan:ignore
		t.Error("a credential-shaped value without an example marker must not be allowlisted")
	}
}

// TestAllowlistDoesNotSuppressRealCredentials guards the direction of error
// that actually matters. A false positive is annoying; an allowlist entry that
// swallows a live key is invisible. Random credentials routinely contain
// substrings like "my-" or "the_", so placeholder patterns must stay anchored.
func TestAllowlistDoesNotSuppressRealCredentials(t *testing.T) {
	allow, err := compileAllowPatterns(defaultAllowPatterns)
	if err != nil {
		t.Fatal(err)
	}

	realKeys := []string{
		"AIzaSyCdmy-K7the_9tSrke72PouQMnMX-a7eZS",  // contains "my-" and "the_" // secscan:ignore
		"ghp_a1B2c3D4e5F6g7H8i9J0k1L2m3N4o5P6q7R8", // secscan:ignore
		"vault_k3Qm9RtZ2xPwL7bNcF4hJ8sVyD6g",
		"AKIAI44QH8DHBEXMPLZZ", // secscan:ignore
	}

	for _, k := range realKeys {
		if isAllowed(k, allow) {
			t.Errorf("%q is a credential-shaped value and must not be allowlisted; "+
				"an over-broad allow pattern silently hides real leaks", k)
		}
	}
}

// TestEntropyTokenizerSplitsStructuredText pins the tokenizer against the
// change that caused the worst noise: `\S{20,}` handed whole URLs and HTML
// attributes to the entropy check as if they were candidate credentials.
func TestEntropyTokenizerSplitsStructuredText(t *testing.T) {
	got := entropyTokens.FindAllString(`href="../user-guide/examples/#share"`, -1)
	for _, tok := range got {
		if len(tok) >= minSecretLen {
			t.Errorf("tokenizer produced %q from a URL: structured text must split "+
				"into short words, not one credential-sized token", tok)
		}
	}

	// A real token must survive tokenising intact.
	const pat = "ghp_a1B2c3D4e5F6g7H8i9J0k1L2m3N4o5P6q7R8" // secscan:ignore
	if toks := entropyTokens.FindAllString(`token = "`+pat+`"`, -1); len(toks) == 0 || toks[0] != pat {
		t.Errorf("tokenizer mangled a real credential: got %v, want [%q]", toks, pat)
	}
}

// TestEntropyRequiresSecretContext pins that entropy is only consulted where a
// line names the value as a credential. Entropy alone is not a detector at any
// threshold: measured across a 37-repo corpus, a 5.0 threshold found no real
// secrets while 4.1 produced over 15,000 findings.
func TestEntropyRequiresSecretContext(t *testing.T) {
	rules, err := compileRules(map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	allow, err := compileAllowPatterns(defaultAllowPatterns)
	if err != nil {
		t.Fatal(err)
	}
	config := &Config{AllowPatterns: allow, EntropyThreshold: defaultEntropy}

	const tok = "ghp_a1B2c3D4e5F6g7H8i9J0k1L2m3N4o5P6q7R8" // secscan:ignore

	if got := scanLine(`const apiToken = "`+tok+`"`, "a.go", 1, rules, config); len(got) == 0 {
		t.Error("a high-entropy token on a line that names it a token should be reported")
	}
	if got := scanLine(`const buildId = "`+tok+`"`, "a.go", 1, rules, config); len(got) != 0 {
		t.Errorf("a high-entropy token with no credential context should not be reported, got %d", len(got))
	}
}

// TestSupabaseAnonKeyIsNotASecret pins that anon keys, which are designed to
// be shipped to browsers and are guarded by row-level security, are dropped --
// while service_role keys, which bypass it, are kept and marked verified.
func TestSupabaseAnonKeyIsNotASecret(t *testing.T) {
	mk := func(role, iss string) string {
		hdr := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
		body := base64.RawURLEncoding.EncodeToString([]byte(
			`{"iss":"` + iss + `","role":"` + role + `"}`))
		return hdr + "." + body + ".Zm9vc2ln"
	}

	got := refineFindings([]Finding{
		{Pattern: "supabase_jwt", RawValue: mk("anon", "supabase")},
		{Pattern: "supabase_jwt", RawValue: mk("service_role", "supabase")},
		{Pattern: "supabase_jwt", RawValue: "eyJnot.a.jwt"},
	})

	if len(got) != 1 {
		t.Fatalf("want only the service_role key reported, got %d findings", len(got))
	}
	if !got[0].Verified {
		t.Error("a service_role key is provably sensitive and should be marked verified")
	}
	if got[0].Confidence < 0.9 {
		t.Errorf("service_role confidence = %.2f, want >= 0.9", got[0].Confidence)
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

// TestGenerateHash verifies that a finding is identified by the secret itself.
//
// This previously asserted the opposite -- that the same secret in a different
// location hashed differently -- which is what made deduplication across
// commits impossible: the history scanner mixed the commit SHA into the key,
// so one leaked credential produced a fresh "unique" finding for every commit
// that touched it.
func TestGenerateHash(t *testing.T) {
	hash1 := generateHash("aws_access_key", "AKIAIOSFODNN7EXAMPLE")
	hash2 := generateHash("aws_access_key", "AKIAIOSFODNN7EXAMPLE")
	hash3 := generateHash("aws_access_key", "AKIAI44QH8DHBEXAMPLE")
	hash4 := generateHash("generic_api_key", "AKIAIOSFODNN7EXAMPLE")

	if hash1 != hash2 {
		t.Error("the same secret should always hash the same")
	}
	if hash1 == hash3 {
		t.Error("different secrets should hash differently")
	}
	if hash1 == hash4 {
		t.Error("the same value under a different rule should hash differently")
	}
	if hash1 == "" {
		t.Error("hash should not be empty")
	}
}

// TestDeduplicateCollapsesSameSecret pins the documented promise that
// duplicate findings are removed across commits.
func TestDeduplicateCollapsesSameSecret(t *testing.T) {
	const secret = "AKIAI44QH8DHBTESTKEY" // secscan:ignore
	h := generateHash("aws_access_key", secret)

	// One secret, seen in the working tree and in three separate commits.
	findings := []Finding{
		{File: "a.js", Line: 1, Pattern: "aws_access_key", RawValue: secret, Hash: h},
		{File: "a.js", Line: 1, Commit: "aaaa", Pattern: "aws_access_key", RawValue: secret, Hash: h},
		{File: "a.js", Line: 9, Commit: "bbbb", Pattern: "aws_access_key", RawValue: secret, Hash: h},
		{File: "a.js", Line: 3, Commit: "cccc", Pattern: "aws_access_key", RawValue: secret, Hash: h},
	}

	got := deduplicateFindings(findings)
	if len(got) != 1 {
		t.Fatalf("one secret across 4 sightings should collapse to 1 finding, got %d", len(got))
	}
	if got[0].Occurrences != 4 {
		t.Errorf("Occurrences = %d, want 4", got[0].Occurrences)
	}
}

// TestDeduplicatePrefersActionableLocation checks that when a secret is live in
// the working tree we report it there, not at some historical commit.
func TestDeduplicatePrefersActionableLocation(t *testing.T) {
	const secret = "AKIAI44QH8DHBTESTKEY" // secscan:ignore
	h := generateHash("aws_access_key", secret)

	got := deduplicateFindings([]Finding{
		{File: "a.js", Commit: "aaaa", Pattern: "aws_access_key", RawValue: secret, Hash: h},
		{File: "a.js", Commit: "", Pattern: "aws_access_key", RawValue: secret, Hash: h},
	})

	if len(got) != 1 {
		t.Fatalf("want 1 finding, got %d", len(got))
	}
	if got[0].Commit != "" {
		t.Error("a secret still present in the working tree should be reported there")
	}
	if got[0].Occurrences != 2 {
		t.Errorf("Occurrences = %d, want 2", got[0].Occurrences)
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

// TestScanGitHistoryHonoursRoot pins the -root flag against the history
// scanner.
//
// The git helpers ran with no Dir set, so they inherited the process working
// directory and ignored -root entirely. `secscan -root /some/repo` -- the
// obvious way to wire this into CI -- reported commits_scanned: 0 and exit 0,
// which is indistinguishable from a clean scan. The headline feature silently
// did nothing.
func TestScanGitHistoryHonoursRoot(t *testing.T) {
	if !gitAvailable() {
		t.Skip("git not available")
	}

	repo := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v (%s)", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "t@example.com")
	run("config", "user.name", "t")

	const secret = "AKIAI44QH8DHBTESTKEY" // secscan:ignore
	if err := os.WriteFile(filepath.Join(repo, "leak.js"),
		[]byte("const k = \""+secret+"\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "-A")
	run("commit", "-qm", "leak")

	rules, err := compileRules(defaultRegexps)
	if err != nil {
		t.Fatal(err)
	}
	allow, err := compileAllowPatterns(defaultAllowPatterns)
	if err != nil {
		t.Fatal(err)
	}
	config := &Config{AllowPatterns: allow, EntropyThreshold: defaultEntropy}
	stats := &Stats{}

	// Deliberately run from a directory that is NOT the repo.
	findings, err := scanGitHistory(repo, rules, config, stats)
	if err != nil {
		t.Fatalf("scanGitHistory(%q) failed: %v", repo, err)
	}
	if stats.CommitsScanned == 0 {
		t.Fatal("scanned 0 commits: -root was ignored and the history scan silently did nothing")
	}

	var found bool
	for _, f := range findings {
		if f.RawValue == secret {
			found = true
			if f.File != "leak.js" {
				t.Errorf("File = %q, want %q: history findings must name the file "+
					"they came from so they are actionable", f.File, "leak.js")
			}
			if f.Line != 1 {
				t.Errorf("Line = %d, want 1: history findings must carry a real "+
					"file line, not an offset into the diff", f.Line)
			}
		}
	}
	if !found {
		t.Error("the committed secret was not found in history")
	}
}

// TestIgnoreDirective pins the per-line opt-out.
//
// A secret scanner's own tests must contain realistic, non-allowlisted
// credentials to prove real ones are not suppressed, so scanning this repo
// reported five of its own fixtures at confidence 0.9. Widening the allowlist
// to cover them would have created silent false negatives -- the one failure
// mode worse than noise. An explicit directive says "I know what this is"
// without weakening detection anywhere else.
func TestIgnoreDirective(t *testing.T) {
	rules, err := compileRules(defaultRegexps)
	if err != nil {
		t.Fatal(err)
	}
	allow, err := compileAllowPatterns(defaultAllowPatterns)
	if err != nil {
		t.Fatal(err)
	}
	config := &Config{AllowPatterns: allow, EntropyThreshold: defaultEntropy}

	line := "key := " + `"AKIAI44QH8DHBEXMPLZZ"` // secscan:ignore

	if got := scanLine(line, "a.go", 1, rules, config); len(got) == 0 {
		t.Fatal("the fixture must be reported without a directive, or this test proves nothing")
	}
	if got := scanLine(line+" // secscan:ignore", "a.go", 1, rules, config); len(got) != 0 {
		t.Errorf("a line marked secscan:ignore must not be reported, got %d findings", len(got))
	}
	if got := scanLine(line+" // secscan: ignore", "a.go", 1, rules, config); len(got) != 0 {
		t.Error("the directive should tolerate a space after the colon")
	}
}

// TestIsExcluded pins .secscanignore matching.
func TestIsExcluded(t *testing.T) {
	globs := []string{"main_test.go", "docs/*", "testdata/"}

	for _, path := range []string{
		"main_test.go",
		"/abs/path/to/main_test.go",
		"docs/quickstart.md",
		"testdata/fixtures.json",
		"a/b/testdata/keys.txt",
	} {
		if !isExcluded(path, globs) {
			t.Errorf("%q should be excluded", path)
		}
	}

	for _, path := range []string{"main.go", "src/app.go", "documentation.md"} {
		if isExcluded(path, globs) {
			t.Errorf("%q must not be excluded", path)
		}
	}

	if isExcluded("main_test.go", nil) {
		t.Error("nothing should be excluded when no ignore file is present")
	}
}

// TestScanGitHistoryRejectsNonRepo pins that a history scan which cannot run
// says so, rather than returning quietly and looking clean.
func TestScanGitHistoryRejectsNonRepo(t *testing.T) {
	if !gitAvailable() {
		t.Skip("git not available")
	}
	_, err := scanGitHistory(t.TempDir(), map[string]*Rule{}, &Config{}, &Stats{})
	if err == nil {
		t.Error("scanning a non-repository should report an error, not silently succeed")
	}
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
