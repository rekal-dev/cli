package scrub

import (
	"math"
	"regexp"
	"strings"
)

const redacted = "[REDACTED]"

// secretPattern pairs a compiled regex with a name for debugging/testing.
type secretPattern struct {
	name string
	re   *regexp.Regexp
}

// patterns is the ordered list of secret-detection regexes.
// Order matters: more specific patterns should come before generic ones.
var patterns = []secretPattern{
	// --- Structured tokens ---
	{"jwt", regexp.MustCompile(`eyJ[A-Za-z0-9_-]{20,}\.eyJ[A-Za-z0-9_-]{20,}\.[A-Za-z0-9_-]{20,}`)},
	{"anthropic_key", regexp.MustCompile(`sk-ant-[A-Za-z0-9_-]{20,}`)},
	{"openai_key", regexp.MustCompile(`sk-[A-Za-z0-9]{20,}`)},
	{"hf_token", regexp.MustCompile(`hf_[A-Za-z0-9]{20,}`)},
	{"github_token", regexp.MustCompile(`(?:ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9]{20,}`)},
	{"github_fine_grained", regexp.MustCompile(`github_pat_[A-Za-z0-9_]{20,}`)},
	{"aws_access_key", regexp.MustCompile(`AKIA[0-9A-Z]{16}`)},
	{"slack_token", regexp.MustCompile(`xox[bpras]-[A-Za-z0-9-]{10,}`)},
	{"npm_token", regexp.MustCompile(`npm_[A-Za-z0-9]{20,}`)},
	{"pypi_token", regexp.MustCompile(`pypi-[A-Za-z0-9_-]{20,}`)},
	{"private_key", regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`)},

	// --- Bearer / Authorization ---
	{"bearer_token", regexp.MustCompile(`(?i)(?:Bearer|Authorization:?\s*Bearer)\s+[A-Za-z0-9_.\-/+=]{20,}`)},

	// --- CLI token flags ---
	{"cli_token_flag", regexp.MustCompile(`(?:--token|--api-key|--secret|--password)[=\s]+\S{8,}`)},

	// --- Database connection strings ---
	{"db_connection", regexp.MustCompile(`(?i)(?:postgres|mysql|mongodb(?:\+srv)?|redis|amqp)://[^\s"'` + "`" + `]{10,}`)},

	// --- Env variable assignments with secret-like names ---
	{"env_secret", regexp.MustCompile(`(?i)(?:^|[\s;])(?:export\s+)?(?:[A-Z_]*(?:SECRET|TOKEN|PASSWORD|API_KEY|APIKEY|ACCESS_KEY|PRIVATE_KEY)[A-Z_]*)=[^\s]{8,}`)},

	// --- Generic secret assignments (key = "value" patterns) ---
	{"generic_secret_assign", regexp.MustCompile(`(?i)(?:secret|token|password|api_key|apikey|access_key|private_key)\s*[:=]\s*["']?[A-Za-z0-9_.\-/+=]{16,}["']?`)},

	// --- AWS secret access key (40 char base64) ---
	{"aws_secret_key", regexp.MustCompile(`(?i)(?:aws_secret_access_key|secret_access_key)\s*[:=]\s*[A-Za-z0-9/+=]{40}`)},

	// --- URL query params with secret-like names ---
	{"url_secret_param", regexp.MustCompile(`(?i)[?&](?:token|key|secret|password|api_key|access_token|auth)=[A-Za-z0-9_.\-/+=]{8,}`)},

	// --- Email addresses ---
	{"email", regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)},

	// --- IP addresses (IPv4) ---
	{"ipv4", regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\b`)},

	// --- High-entropy strings (handled specially) ---
	{"high_entropy", regexp.MustCompile(`(?:^|[^A-Za-z0-9_.\-/])([A-Za-z0-9_.\-/+=]{40,})(?:$|[^A-Za-z0-9_.\-/])`)},
}

// allowedEmails are email addresses that should not be redacted.
var allowedEmails = map[string]bool{
	"noreply@anthropic.com":  true,
	"noreply@github.com":     true,
	"noreply@example.com":    true,
	"user@example.com":       true,
	"test@example.com":       true,
	"example@example.com":    true,
	"git@github.com":         true,
	"actions@github.com":     true,
	"dependabot@github.com":  true,
	"bot@renovateapp.com":    true,
	"support@github.com":     true,
}

// allowedEmailDomains are domains whose emails should not be redacted.
var allowedEmailDomains = map[string]bool{
	"example.com":     true,
	"example.org":     true,
	"example.net":     true,
	"test.com":        true,
	"localhost":        true,
}

// privateIPPrefixes are IP prefixes that should not be redacted.
var privateIPPrefixes = []string{
	"10.", "172.16.", "172.17.", "172.18.", "172.19.",
	"172.20.", "172.21.", "172.22.", "172.23.",
	"172.24.", "172.25.", "172.26.", "172.27.",
	"172.28.", "172.29.", "172.30.", "172.31.",
	"192.168.", "127.", "0.0.0.0", "255.255.255.255",
}

// decoratorPattern matches Python decorator-like patterns that look like high-entropy strings.
var decoratorPattern = regexp.MustCompile(`^@[A-Za-z]`)

// commonPathPattern matches file paths that look like high-entropy strings.
var commonPathPattern = regexp.MustCompile(`^[/.]|[/\\]`)

// shannonEntropy computes the Shannon entropy of a string in bits per character.
func shannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	freq := make(map[rune]float64)
	for _, c := range s {
		freq[c]++
	}
	n := float64(len([]rune(s)))
	var entropy float64
	for _, count := range freq {
		p := count / n
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}
	return entropy
}

// isAllowedEmail returns true if the email should not be redacted.
func isAllowedEmail(email string) bool {
	lower := strings.ToLower(email)
	if allowedEmails[lower] {
		return true
	}
	parts := strings.SplitN(lower, "@", 2)
	if len(parts) == 2 && allowedEmailDomains[parts[1]] {
		return true
	}
	return false
}

// isPrivateIP returns true if the IP address is private/reserved.
func isPrivateIP(ip string) bool {
	for _, prefix := range privateIPPrefixes {
		if strings.HasPrefix(ip, prefix) {
			return true
		}
	}
	return false
}

// isAllowedHighEntropy returns true if a high-entropy match is a false positive.
func isAllowedHighEntropy(s string) bool {
	// Decorator patterns (@Component, etc.)
	if decoratorPattern.MatchString(s) {
		return true
	}
	// File paths
	if commonPathPattern.MatchString(s) {
		return true
	}
	// Hex strings that are likely hashes (git SHAs, etc.) — allow if all hex
	if isHexString(s) && len(s) <= 64 {
		return true
	}
	return false
}

func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

const highEntropyThreshold = 4.5

// RedactText scans text for secrets and replaces matches with [REDACTED].
func RedactText(text string) string {
	for _, p := range patterns {
		switch p.name {
		case "email":
			text = p.re.ReplaceAllStringFunc(text, func(match string) string {
				if isAllowedEmail(match) {
					return match
				}
				return redacted
			})
		case "ipv4":
			text = p.re.ReplaceAllStringFunc(text, func(match string) string {
				if isPrivateIP(match) {
					return match
				}
				return redacted
			})
		case "high_entropy":
			text = p.re.ReplaceAllStringFunc(text, func(match string) string {
				// The regex captures with surrounding context; extract the group.
				sub := p.re.FindStringSubmatch(match)
				if len(sub) < 2 {
					return match
				}
				candidate := sub[1]
				if isAllowedHighEntropy(candidate) {
					return match
				}
				if shannonEntropy(candidate) < highEntropyThreshold {
					return match
				}
				return strings.Replace(match, candidate, redacted, 1)
			})
		default:
			text = p.re.ReplaceAllString(text, redacted)
		}
	}
	return text
}
