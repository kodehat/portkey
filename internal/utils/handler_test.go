package utils

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNotFoundRespWrWriteHeader200(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &NotFoundRespWr{ResponseWriter: rec}
	w.WriteHeader(http.StatusOK)
	if w.status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.status)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected recorder code %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestNotFoundRespWrWriteHeader404(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &NotFoundRespWr{ResponseWriter: rec}
	w.WriteHeader(http.StatusNotFound)
	if w.status != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.status)
	}
}

func TestNotFoundRespWrWritePassthrough(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &NotFoundRespWr{ResponseWriter: rec}
	w.WriteHeader(http.StatusOK)
	n, err := w.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatalf("expected written 5 bytes, got %d", n)
	}
	if rec.Body.String() != "hello" {
		t.Fatalf("expected body 'hello', got %q", rec.Body.String())
	}
}

func TestNotFoundRespWrWriteSuppressed(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &NotFoundRespWr{ResponseWriter: rec}
	w.WriteHeader(http.StatusNotFound)
	n, err := w.Write([]byte("not found"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 9 {
		t.Fatalf("expected reported 9 bytes, got %d", n)
	}
	if rec.Body.String() != "" {
		t.Fatalf("expected empty body, got %q", rec.Body.String())
	}
}

func TestStaticHandlerFound(t *testing.T) {
	if _, err := os.Stat("testdata/static.css"); err != nil {
		t.Skip("testdata/static.css does not exist")
	}
	fs := http.FileServer(http.Dir("testdata"))
	h := StaticHandler(fs)
	req := httptest.NewRequest(http.MethodGet, "/static.css", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestStaticHandlerMissing(t *testing.T) {
	fs := http.FileServer(http.Dir("testdata"))
	h := StaticHandler(fs)
	req := httptest.NewRequest(http.MethodGet, "/does-not-exist-xyz.css", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect for missing file, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "/404" {
		t.Fatalf("expected Location /404, got %q", loc)
	}
}
