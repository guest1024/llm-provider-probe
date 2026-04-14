package runner

import "testing"

func TestSanitizeTextRedactsBearerAndVendorKeys(t *testing.T) {
	in := "Authorization: Bearer ms-11111111-2222-3333-4444-555555555555 and sk-testabcdefghijklmnopqrstuvwxyz123456"
	got := sanitizeText(in)
	if got == in {
		t.Fatalf("expected redaction, got unchanged: %s", got)
	}
	if contains(got, "11111111-2222") || contains(got, "sk-testabcdefghijklmnopqrstuvwxyz") {
		t.Fatalf("token leaked after redaction: %s", got)
	}
}

func TestSanitizeHeadersRedactsSensitiveHeaders(t *testing.T) {
	got := sanitizeHeaders(map[string][]string{
		"Authorization": {"Bearer abcdefghijklmnop"},
		"X-Trace":       {"token=abcdefghijklmnop"},
	})
	if got["Authorization"][0] != "[REDACTED]" {
		t.Fatalf("expected authorization redacted, got %#v", got["Authorization"])
	}
	if got["X-Trace"][0] == "token=abcdefghijklmnop" {
		t.Fatalf("expected header value sanitized, got %#v", got["X-Trace"])
	}
}

func contains(s, sub string) bool {
	return len(sub) > 0 && (len(s) >= len(sub)) && (stringIndex(s, sub) >= 0)
}
func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
