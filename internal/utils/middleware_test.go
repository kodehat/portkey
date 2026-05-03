package utils

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
)

func TestIpFromRequestPlain(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1"
	ip, err := ipFromRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	if ip.String() != "192.168.1.1" {
		t.Fatalf("expected 192.168.1.1, got %q", ip.String())
	}
}

func TestIpFromRequestWithPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:8080"
	ip, err := ipFromRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	if ip.String() != "192.168.1.1" {
		t.Fatalf("expected 192.168.1.1, got %q", ip.String())
	}
}

func TestIpFromRequestIPv6(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "[::1]:8080"
	ip, err := ipFromRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	if ip.String() != "::1" {
		t.Fatalf("expected ::1, got %q", ip.String())
	}
}

func TestIpRangesFromString(t *testing.T) {
	prefixes := ipRangesFromString([]string{"10.0.0.0/8", "192.168.0.0/16"})
	if len(prefixes) != 2 {
		t.Fatalf("expected 2 prefixes, got %d", len(prefixes))
	}
	if !prefixes[0].Contains(mustParseAddr("10.1.2.3")) {
		t.Fatal("expected 10.1.2.3 to be in 10.0.0.0/8")
	}
	if !prefixes[1].Contains(mustParseAddr("192.168.1.1")) {
		t.Fatal("expected 192.168.1.1 to be in 192.168.0.0/16")
	}
}

func TestIpCheckAllowed(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := IpCheck(logger, []string{"127.0.0.1/32"}, handler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1"
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestIpCheckDenied(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called")
	})
	wrapped := IpCheck(logger, []string{"10.0.0.0/8"}, handler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1"
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestIpCheckParseError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called")
	})
	wrapped := IpCheck(logger, []string{"127.0.0.1/32"}, handler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "not-an-ip"
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func mustParseAddr(s string) netip.Addr {
	addr, err := netip.ParseAddr(s)
	if err != nil {
		panic(err)
	}
	return addr
}
