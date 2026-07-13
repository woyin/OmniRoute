package main

import "testing"

func TestValidateWebhookURL(t *testing.T) {
	t.Setenv("OMNIROUTE_ALLOW_PRIVATE_PROVIDER_URLS", "")
	t.Setenv("OUTBOUND_SSRF_GUARD_ENABLED", "")
	cases := []struct {
		url, reason string
		valid       bool
	}{
		{"https://example.com/hook", "", true},
		{"ftp://example.com/hook", "invalid_url", false},
		{"https://user:pass@example.com", "blocked_private", false},
		{"http://127.0.0.1/hook", "blocked_private", false},
		{"http://169.254.169.254/latest/meta-data", "blocked_private", false},
	}
	for _, tc := range cases {
		valid, reason := validateWebhookURL(tc.url)
		if valid != tc.valid || reason != tc.reason {
			t.Errorf("%s: got (%v,%q), want (%v,%q)", tc.url, valid, reason, tc.valid, tc.reason)
		}
	}
	t.Setenv("OMNIROUTE_ALLOW_PRIVATE_PROVIDER_URLS", "true")
	if valid, _ := validateWebhookURL("http://127.0.0.1/hook"); !valid {
		t.Error("private opt-in should allow loopback webhook")
	}
	if valid, _ := validateWebhookURL("http://169.254.169.254/latest/meta-data"); valid {
		t.Error("private opt-in must never allow metadata endpoint")
	}
}
