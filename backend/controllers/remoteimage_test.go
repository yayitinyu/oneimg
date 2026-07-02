package controllers

import (
	"net"
	"testing"
)

func TestIsPublicIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{name: "public IPv4", ip: "8.8.8.8", want: true},
		{name: "loopback", ip: "127.0.0.1", want: false},
		{name: "private IPv4", ip: "192.168.1.10", want: false},
		{name: "link local", ip: "169.254.169.254", want: false},
		{name: "IPv6 loopback", ip: "::1", want: false},
		{name: "IPv6 private", ip: "fd00::1", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPublicIP(net.ParseIP(tt.ip)); got != tt.want {
				t.Fatalf("isPublicIP(%q) = %t, want %t", tt.ip, got, tt.want)
			}
		})
	}
}

func TestValidateRemoteURLRejectsUnsafeTargets(t *testing.T) {
	tests := []string{
		"file:///etc/passwd",
		"http://127.0.0.1/image.png",
		"http://[::1]/image.png",
		"http://user:pass@8.8.8.8/image.png",
	}
	for _, rawURL := range tests {
		if _, err := validateRemoteURL(t.Context(), rawURL); err == nil {
			t.Fatalf("validateRemoteURL(%q) unexpectedly succeeded", rawURL)
		}
	}
}

func TestNormalizeContentType(t *testing.T) {
	if got := normalizeContentType(" Image/PNG; charset=binary "); got != "image/png" {
		t.Fatalf("normalizeContentType() = %q", got)
	}
}

func TestSafeLocalUploadPath(t *testing.T) {
	if _, ok := safeLocalUploadPath("/uploads/2026/07/image.webp"); !ok {
		t.Fatal("expected generated upload path to be accepted")
	}
	for _, unsafePath := range []string{
		"/etc/passwd",
		"/uploads/../../.env",
		"/uploads/..\\..\\.env",
	} {
		if path, ok := safeLocalUploadPath(unsafePath); ok {
			t.Fatalf("unsafe path %q accepted as %q", unsafePath, path)
		}
	}
}

func TestStorageObjectKey(t *testing.T) {
	tests := map[string]string{
		"/uploads/2026/07/image.webp":                    "uploads/2026/07/image.webp",
		"https://cdn.example.com/uploads/2026/07/a.webp": "uploads/2026/07/a.webp",
	}
	for input, want := range tests {
		if got := storageObjectKey(input); got != want {
			t.Fatalf("storageObjectKey(%q) = %q, want %q", input, got, want)
		}
	}
}
