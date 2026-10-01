package external

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const page = `<!DOCTYPE html><html><body>
<div id="body" class="width">
<h1>ARTI NAMA</h1>
<script>(adsbygoogle = window.adsbygoogle || []).push({});</script>
Nama Viky, memiliki arti: Pemimpin.<br>Kalimat dua.<br>
<img border="0" src="ramalan_kecocokan_cinta4.png"><br><br>Saran &amp; nasihat<a href="x">&lt; Hitung Kembali</a>
<font><i>Tidak ditemukan</i></font>
<span><img src="rejeki/15.png"></span>
</div>
<div class="main-horoscope"><p>Oct 1, 2026 - Today is good.</p><p>More Horoscopes for Leo</p></div>
<div class="grid"><span>A</span></div><section class="grid x">B</section>
</body></html>`

func TestFindAndText(t *testing.T) {
	doc, err := ParseHTML(page)
	if err != nil {
		t.Fatal(err)
	}
	body := doc.Find("#body").Text()
	if !strings.Contains(body, "(adsbygoogle = window.adsbygoogle || []).push({});") {
		t.Error("script text must be included, like cheerio .text()")
	}
	if !strings.Contains(body, "Saran & nasihat") {
		t.Error("entities must be decoded in .text()")
	}
	if got := doc.Find("#body > font > i").Text(); got != "Tidak ditemukan" {
		t.Errorf("child selector: %q", got)
	}
	if got := doc.Find(".main-horoscope > p").Text(); got != "Oct 1, 2026 - Today is good.More Horoscopes for Leo" {
		t.Errorf("multiple matches concatenate: %q", got)
	}
	if got := doc.Find(".grid").Text(); got != "AB" {
		t.Errorf("class selector: %q", got)
	}
	if src, ok := doc.Find("#body > span > img").Attr("src"); !ok || src != "rejeki/15.png" {
		t.Errorf("attr: %q %v", src, ok)
	}
	if _, ok := doc.Find("#body > img.nope").Attr("src"); ok {
		t.Error("attr on empty selection must report false")
	}
}

func TestInnerHTMLMatchesParse5(t *testing.T) {
	doc, _ := ParseHTML(page)
	h, ok := doc.Find("#body").HTML()
	if !ok {
		t.Fatal("no #body")
	}
	for _, want := range []string{
		`<br>Kalimat dua.<br>`,                                // void elements without a slash
		`<img border="0" src="ramalan_kecocokan_cinta4.png">`, // attribute order kept
		`png"><br><br>Saran &amp; nasihat`,                    // text re-escaped
		`&lt; Hitung Kembali`,                                 // < escaped in text
		`push({});</script>`,                                  // script text raw
	} {
		if !strings.Contains(h, want) {
			t.Errorf("inner HTML missing %q in\n%s", want, h)
		}
	}
}

func TestStrPipeline(t *testing.T) {
	got, err := Str("a-b-c").Split("-", 1).Value()
	if err != nil || got != "b" {
		t.Errorf("Split: %q %v", got, err)
	}
	if _, err := Str("abc").Split("-", 1).Replace("a", "b").Value(); err != ErrUnexpectedContent {
		t.Errorf("missing index must fail the chain, got %v", err)
	}
	if got := Str("x.x.x").Replace(".", "!").String(); got != "x!x.x" {
		t.Errorf("Replace replaces the first match only: %q", got)
	}
	if got := Str("\n\n  a\n \n\nb").ReplaceReAll(ReBlankLines, "").String(); got != "  a\nb" {
		t.Errorf("ReBlankLines: %q", got)
	}
	if got := Str("x < Hitung Kembali\nmore").ReplaceRe(ReHitungKembali, "").String(); got != "x " {
		t.Errorf("ReHitungKembali: %q", got)
	}
	if got := Str("  a ").TrimStart().TrimEnd().String(); got != "a" {
		t.Errorf("trim: %q", got)
	}
}

func TestTranslate(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("q")
		if r.URL.Query().Get("sl") != "id" || r.URL.Query().Get("tl") != "en" {
			t.Errorf("languages: %v", r.URL.Query())
		}
		w.Write([]byte(`{"sentences":[{"trans":"Hello. ","orig":"Halo. "},{"trans":"World"},{"src_translit":"x"}]}`))
	}))
	defer srv.Close()
	tr := &Translator{Client: NewClient(5 * time.Second), Base: srv.URL}
	got, err := tr.Translate(context.Background(), "Halo & #dunia", "id", "en")
	if err != nil {
		t.Fatal(err)
	}
	if gotQuery != "Halo dan #dunia" {
		t.Errorf("query = %q (& must become dan, # must survive encoding)", gotQuery)
	}
	if len(got) != 2 || got[0] != "Hello. " || got[1] != "World" {
		t.Errorf("sentences = %q", got)
	}
}

func TestClientRejectsErrorsAndCapsBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/fail" {
			http.Error(w, "nope", http.StatusBadGateway)
			return
		}
		if r.Method == http.MethodPost && r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("content type %q", r.Header.Get("Content-Type"))
		}
		w.Write([]byte(strings.Repeat("x", 100)))
	}))
	defer srv.Close()
	c := NewClient(5 * time.Second)
	c.MaxBody = 10
	if _, err := c.Get(context.Background(), srv.URL+"/fail"); err == nil {
		t.Error("expected error for 502")
	}
	body, err := c.PostForm(context.Background(), srv.URL+"/ok", map[string][]string{"a": {"b"}})
	if err != nil || len(body) != 10 {
		t.Errorf("body len %d err %v", len(body), err)
	}
}
