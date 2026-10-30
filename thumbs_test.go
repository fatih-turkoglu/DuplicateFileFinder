package main

import (
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestThumbMiddleware(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "foto klasörü", "a.png")
	os.MkdirAll(filepath.Dir(imgPath), 0o755)
	f, _ := os.Create(imgPath)
	png.Encode(f, image.NewRGBA(image.Rect(0, 0, 400, 200)))
	f.Close()
	txtPath := filepath.Join(dir, "gizli.txt")
	os.WriteFile(txtPath, []byte("okunmamalı"), 0o644)

	// Taranmış ama diskte olmayan dosya ile taranmamış ama diskte duran görsel.
	missing := filepath.Join(dir, "yok.png")
	unscanned := filepath.Join(dir, "taranmadi.png")
	f, _ = os.Create(unscanned)
	png.Encode(f, image.NewRGBA(image.Rect(0, 0, 10, 10)))
	f.Close()
	indexed := func(p string) bool { return p == imgPath || p == txtPath || p == missing }

	nextCalled := false
	h := thumbMiddleware(newThumbCache(1<<20), indexed)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	}))
	get := func(target string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
		return rec
	}
	thumb := func(p string) string { return "/thumb?s=100&path=" + url.QueryEscape(p) }

	rec := get(thumb(imgPath))
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "image/jpeg" || rec.Body.Len() == 0 {
		t.Fatalf("görsel: %d %s (%d bayt)", rec.Code, rec.Header().Get("Content-Type"), rec.Body.Len())
	}
	if rec := get(thumb(txtPath)); rec.Code != http.StatusBadRequest {
		t.Errorf("görsel olmayan dosya: %d, beklenen 400", rec.Code)
	}
	if rec := get(thumb("goreli.png")); rec.Code != http.StatusBadRequest {
		t.Errorf("göreli yol: %d, beklenen 400", rec.Code)
	}
	if rec := get(thumb(missing)); rec.Code != http.StatusNotFound {
		t.Errorf("olmayan dosya: %d, beklenen 404", rec.Code)
	}
	if rec := get(thumb(unscanned)); rec.Code != http.StatusNotFound {
		t.Errorf("taranmamış dosya: %d, beklenen 404", rec.Code)
	}
	if get("/index.html"); !nextCalled {
		t.Error("diğer istekler sonraki işleyiciye geçmeli")
	}
}

func TestThumbCacheEviction(t *testing.T) {
	c := newThumbCache(10)
	c.put("a", make([]byte, 4))
	c.put("b", make([]byte, 4))
	c.get("a") // a en yeni
	c.put("c", make([]byte, 4))
	if _, ok := c.get("b"); ok {
		t.Error("en uzun süredir kullanılmayan (b) atılmalıydı")
	}
	if _, ok := c.get("a"); !ok {
		t.Error("a atılmamalıydı")
	}
	if c.bytes > 10 {
		t.Errorf("sınır aşıldı: %d", c.bytes)
	}
}
