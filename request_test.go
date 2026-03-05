package web

import "testing"

func TestContainsOriginStrictMatch(t *testing.T) {
	if containsOrigin([]string{"https://good.com"}, "https://good.com.evil.com") {
		t.Fatalf("expected strict origin check to reject prefix match")
	}

	if !containsOrigin([]string{"https://good.com"}, "https://good.com") {
		t.Fatalf("expected exact origin to pass")
	}
}

func TestContainsOriginWildcard(t *testing.T) {
	if !containsOrigin([]string{"https://*.good.com"}, "https://api.good.com") {
		t.Fatalf("expected wildcard subdomain to pass")
	}

	if containsOrigin([]string{"https://*.good.com"}, "https://good.com") {
		t.Fatalf("expected bare root domain to fail wildcard subdomain match")
	}
}
