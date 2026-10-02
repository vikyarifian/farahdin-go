package service

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/vikyarifian/farahdin-go/internal/domain"
	"github.com/vikyarifian/farahdin-go/internal/external"
	"github.com/vikyarifian/farahdin-go/internal/repository"
)

// fakeFetcher returns canned pages by URL prefix and records requests.
type fakeFetcher struct {
	pages map[string]string
	calls []string
	forms []url.Values
}

func (f *fakeFetcher) lookup(u string) (string, error) {
	f.calls = append(f.calls, u)
	best := ""
	for prefix := range f.pages {
		if strings.HasPrefix(u, prefix) && len(prefix) > len(best) {
			best = prefix
		}
	}
	if best == "" {
		return "", errors.New("no fixture for " + u)
	}
	return f.pages[best], nil
}

func (f *fakeFetcher) Get(_ context.Context, u string) (string, error) { return f.lookup(u) }

func (f *fakeFetcher) PostForm(_ context.Context, u string, form url.Values) (string, error) {
	f.forms = append(f.forms, form)
	return f.lookup(u)
}

// fakeTranslator tags text instead of translating it.
type fakeTranslator struct{ calls int }

func (f *fakeTranslator) Translate(_ context.Context, text, from, to string) ([]string, error) {
	f.calls++
	return []string{"[" + from + ">" + to + "] " + text}, nil
}

var fixedNow = time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)

func newReadings(pages map[string]string) (*Readings, *fakeFetcher, *fakeTranslator) {
	f := &fakeFetcher{pages: pages}
	tr := &fakeTranslator{}
	src := external.Sources{
		Primbon: "https://p", PrimbonPost: "https://pp", Horoscope: "https://h",
		CaliforniaPsychic: "https://c", MatrixDestiny: "https://m",
	}
	return &Readings{Fetch: f, Translate: tr, Src: src, Now: func() time.Time { return fixedNow }}, f, tr
}

const ads = "<script>(adsbygoogle = window.adsbygoogle || []).push({});</script>"

func TestPrimbonNameMeaning(t *testing.T) {
	r, f, tr := newReadings(map[string]string{
		"https://p/arti_nama.php": `<div id="body">ARTI NAMA` + ads + "\nNama Viky, memiliki arti: Pemimpin.\nKalimat dua.\nNama: <form></form></div>",
	})
	got, err := r.Primbon(context.Background(), 1, PrimbonInput{Name: "Viky A"}, "ID")
	if err != nil {
		t.Fatal(err)
	}
	// The trailing "" comes from the source's newline split; CleanLines drops it when rendering.
	if strings.Join(got.Lines, "|") != "Nama Viky, memiliki arti: Pemimpin.|Kalimat dua.|" {
		t.Errorf("lines = %q", got.Lines)
	}
	if tr.calls != 0 {
		t.Error("ID result must not call the translator")
	}
	if !strings.Contains(f.calls[0], "nama1=Viky+A") {
		t.Errorf("name must be URL-encoded: %s", f.calls[0])
	}

	got, err = r.Primbon(context.Background(), 1, PrimbonInput{Name: "Viky"}, "EN")
	if err != nil || len(got.Lines) != 1 || !strings.HasPrefix(got.Lines[0], "[id>en] Nama Viky") {
		t.Errorf("EN = %q, %v", got.Lines, err)
	}
}

func TestPrimbonMatchLoveMeter(t *testing.T) {
	body := `<div id="body">` + ads + `Nama Anda: Viky<br>Pasangan: Farah<br>Sisi Positif Anda: Kuat.<br>Sisi Negatif Anda: Keras kepala.<br>` +
		`<img border="0" src="ramalan_kecocokan_cinta4.png"><br><br>Hubungan yang aneh.<a href="x">&lt; Hitung Kembali</a></div>`
	r, _, _ := newReadings(map[string]string{"https://p/kecocokan_nama_pasangan.php": body})
	got, err := r.Primbon(context.Background(), 3, PrimbonInput{Name: "Viky", Partner: "Farah"}, "ID")
	if err != nil {
		t.Fatal(err)
	}
	if got.Love != 4 {
		t.Errorf("love = %d, want 4", got.Love)
	}
	want := []string{"Nama Anda: Viky", "Pasangan: Farah", "", "Sisi Positif Anda: Kuat.. ", "Sisi Negatif Anda:  Keras kepala.. ", "Hubungan yang aneh."}
	if strings.Join(got.Lines, "|") != strings.Join(want, "|") {
		t.Errorf("lines =\n%q\nwant\n%q", got.Lines, want)
	}
}

func TestPrimbonValidation(t *testing.T) {
	r, f, _ := newReadings(nil)
	_, err := r.Primbon(context.Background(), 3, PrimbonInput{Name: "Viky"}, "EN")
	if msg, ok := IsValidation(err); !ok || msg != "Partner name is required." {
		t.Errorf("err = %v", err)
	}
	if len(f.calls) != 0 {
		t.Error("invalid input must not reach upstream")
	}
}

func TestPrimbonGoodDayUsesBirthday(t *testing.T) {
	r, f, _ := newReadings(map[string]string{"https://pp/petung_hari_baik.php": `<div id="body">PETUNG Kamarokam Jumat Paing, Tgl. 17 Agustus 1990 Baik</div>`})
	b := time.Date(1990, 8, 17, 0, 0, 0, 0, time.UTC)
	d := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	if _, err := r.Primbon(context.Background(), 8, PrimbonInput{Birthday: b, Date: d}, "ID"); err != nil {
		t.Fatal(err)
	}
	if f.forms[0].Get("thn") != "1990" || f.forms[0].Get("tgl") != "17" {
		t.Errorf("form = %v (source parity: birthday, not the Date field)", f.forms[0])
	}
}

func TestHoroscopeToday(t *testing.T) {
	r, f, _ := newReadings(map[string]string{
		"https://h/us/horoscopes/general/horoscope-general-daily-today.aspx": `<div class="main-horoscope"><p><strong>Oct 1, 2026</strong> - Bold moves pay off.Reveal what 2026 has in the stars for you with your Yearly Horoscope!</p><p>More Horoscopes for Leo</p></div>`,
	})
	in := HoroscopeInput{Birthday: time.Date(1990, 8, 17, 0, 0, 0, 0, time.UTC)}
	got, err := r.Horoscope(context.Background(), 2, in, "EN")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(f.calls[0], "?sign=5") {
		t.Errorf("Leo must be sign 5: %s", f.calls[0])
	}
	if len(got.Lines) != 1 || got.Lines[0] != "Bold moves pay off." {
		t.Errorf("lines = %q", got.Lines)
	}
}

func TestTarotCardImages(t *testing.T) {
	cases := map[int]string{1: "/static/images/tarot/1.png", 22: "/static/images/tarot/22.png",
		-1: "/static/images/tarot/your-card.png", -2: "/static/images/tarot/partner-card.png", 25: "/static/images/tarot/tarot.png"}
	for in, want := range cases {
		if got := TarotCardImage(in); got != want {
			t.Errorf("TarotCardImage(%d) = %q", in, got)
		}
	}
	if got := TarotExtraImage(4, 25); got != "/static/images/tarot/tarot.png" {
		t.Errorf("past life >22 = %q", got)
	}
	if got := TarotExtraImage(1, 3); got != "" {
		t.Errorf("love has no extra image: %q", got)
	}
}

func TestShuffle(t *testing.T) {
	cards := Shuffle(22, 30)
	seen := map[int]bool{}
	for _, c := range cards {
		if c < 1 || c > 30 || seen[c] {
			t.Fatalf("bad deal %v", cards)
		}
		seen[c] = true
	}
	if len(cards) != 22 {
		t.Errorf("len = %d", len(cards))
	}
}

func TestMatrixDestinyReadingFailureKeepsChart(t *testing.T) {
	r, _, _ := newReadings(map[string]string{"https://m/": "<html>no nonce here</html>"})
	res, err := r.MatrixDestiny(context.Background(), "viky arifian", time.Date(1990, 8, 17, 0, 0, 0, 0, time.UTC), "EN")
	if err != nil {
		t.Fatal(err)
	}
	if res.Name != "Viky Arifian" || res.Matrix.Points["apoint"] != 17 || res.Reading.Notice == "" {
		t.Errorf("result = %+v", res)
	}
}

func TestMatrixDestinyReading(t *testing.T) {
	ajax := `{"infg1":"Calm<br />","infg2":"Bright","infg3":"Karma","infg4":"Love","infg5":"Money","infg6":"Talent","infg8":"Purpose"}`
	r, f, _ := newReadings(map[string]string{
		"https://m/wp-admin/admin-ajax.php": ajax,
		"https://m/":                        `<script>var ajax_var = {"url":"x","nonce":"abc123"};</script>`,
	})
	res, err := r.MatrixDestiny(context.Background(), "Viky", time.Date(1990, 8, 17, 0, 0, 0, 0, time.UTC), "EN")
	if err != nil {
		t.Fatal(err)
	}
	if f.forms[0].Get("nonce_code") != "abc123" || f.forms[0].Get("yourbirthday") != "17/08/1990" {
		t.Errorf("form = %v", f.forms[0])
	}
	if res.Reading.Lines[0] != "1. Soul comfort" || res.Reading.Lines[1] != "Calm" || len(res.Reading.Lines) != 14 {
		t.Errorf("lines = %q", res.Reading.Lines)
	}
}

func TestMatrixValidation(t *testing.T) {
	r, _, _ := newReadings(nil)
	_, err := r.MatrixDestiny(context.Background(), "", time.Time{}, "EN")
	if _, ok := IsValidation(err); !ok {
		t.Errorf("err = %v", err)
	}
}

func TestCleanLines(t *testing.T) {
	got := CleanLines([]string{"  a", ".", " . ", "", "<br>", "b "})
	if strings.Join(got, "|") != "a|b " {
		t.Errorf("CleanLines = %q", got)
	}
}

// --- profiles ---

type memUsers struct {
	byID   map[int64]*domain.User
	nextID int64
}

func newMemUsers() *memUsers { return &memUsers{byID: map[int64]*domain.User{}} }

func (m *memUsers) find(pred func(*domain.User) bool) (*domain.User, error) {
	for id := int64(1); id <= m.nextID; id++ {
		if u, ok := m.byID[id]; ok && pred(u) {
			c := *u
			return &c, nil
		}
	}
	return nil, repository.ErrNotFound
}
func (m *memUsers) ByID(_ context.Context, id int64) (*domain.User, error) {
	return m.find(func(u *domain.User) bool { return u.ID == id })
}
func (m *memUsers) ByGoogleSub(_ context.Context, sub string) (*domain.User, error) {
	return m.find(func(u *domain.User) bool { return u.GoogleSub == sub })
}
func (m *memUsers) FirstByEmail(_ context.Context, email string) (*domain.User, error) {
	return m.find(func(u *domain.User) bool { return u.Email == email })
}
func (m *memUsers) Insert(_ context.Context, u *domain.User) (int64, error) {
	m.nextID++
	c := *u
	c.ID = m.nextID
	m.byID[c.ID] = &c
	return c.ID, nil
}
func (m *memUsers) LinkIdentity(_ context.Context, id int64, username, fullname, sub, img string) error {
	u := m.byID[id]
	u.Username, u.Fullname, u.GoogleSub, u.ImageURL = username, fullname, sub, img
	return nil
}
func (m *memUsers) SetImage(_ context.Context, id int64, img string) error {
	m.byID[id].ImageURL = img
	return nil
}
func (m *memUsers) UpdateProfile(_ context.Context, id int64, fullname, birthday, birthplace, gender, zodiac string) error {
	u := m.byID[id]
	u.Fullname, u.Birthday, u.Birthplace, u.Gender, u.Zodiac = fullname, birthday, birthplace, gender, zodiac
	return nil
}
func (m *memUsers) UpdateBirthday(_ context.Context, id int64, birthday, zodiac string) error {
	u := m.byID[id]
	u.Birthday, u.Zodiac = birthday, zodiac
	return nil
}

func TestSignInPortsCreateUser(t *testing.T) {
	users := newMemUsers()
	p := &Profiles{Users: users, Now: func() time.Time { return fixedNow }}
	ctx := context.Background()

	// New user: defaults from createUser.
	u, err := p.SignIn(ctx, domain.Identity{Subject: "g1", Email: "viky@example.com", FirstName: "Viky", LastName: "Arifian"})
	if err != nil {
		t.Fatal(err)
	}
	if u.Username != "viky" || u.Fullname != "Viky Arifian" || u.Birthday != "2026-10-01" || u.Zodiac != "" || u.Gender != "" {
		t.Errorf("new user = %+v", u)
	}

	// Same subject: returned unchanged.
	again, _ := p.SignIn(ctx, domain.Identity{Subject: "g1", Email: "other@example.com", FirstName: "X"})
	if again.ID != u.ID || again.Fullname != "Viky Arifian" {
		t.Errorf("existing subject must not be patched: %+v", again)
	}

	// Imported user (e.g. from Convex) matched by email: username, fullname, subject patched.
	users.Insert(ctx, &domain.User{Username: "old", Fullname: "Old Name", Email: "farah@example.com", Birthday: "1992-03-03", Zodiac: "Pisces"})
	linked, err := p.SignIn(ctx, domain.Identity{Subject: "g2", Email: "farah@example.com", FirstName: "Farah"})
	if err != nil {
		t.Fatal(err)
	}
	if linked.Username != "farah" || linked.Fullname != "Farah" || linked.GoogleSub != "g2" || linked.Birthday != "1992-03-03" {
		t.Errorf("linked = %+v", linked)
	}
}

func TestUpdateProfile(t *testing.T) {
	users := newMemUsers()
	p := &Profiles{Users: users, Now: func() time.Time { return fixedNow }}
	ctx := context.Background()
	u, _ := p.SignIn(ctx, domain.Identity{Subject: "g1", Email: "viky@example.com", FirstName: "Viky"})

	errs, err := p.UpdateProfile(ctx, u, domain.ProfileUpdate{Fullname: " Viky A ", Birthday: time.Date(1990, 8, 17, 0, 0, 0, 0, time.UTC), Birthplace: "Jakarta", Gender: "Male"}, "EN")
	if err != nil || len(errs) != 0 {
		t.Fatalf("errs=%v err=%v", errs, err)
	}
	saved, _ := users.ByID(ctx, u.ID)
	if saved.Fullname != "Viky A" || saved.Zodiac != "Leo" || saved.Email != "viky@example.com" {
		t.Errorf("saved = %+v", saved)
	}

	errs, _ = p.UpdateProfile(ctx, u, domain.ProfileUpdate{Birthday: fixedNow.AddDate(0, 0, 2), Gender: "Other"}, "EN")
	if errs["birthday"] == "" || errs["gender"] == "" {
		t.Errorf("errs = %v", errs)
	}
}

func TestSyncBirthday(t *testing.T) {
	users := newMemUsers()
	p := &Profiles{Users: users, Now: func() time.Time { return fixedNow }}
	ctx := context.Background()
	u, _ := p.SignIn(ctx, domain.Identity{Subject: "g1", Email: "viky@example.com"})
	if err := p.SyncBirthday(ctx, u, time.Date(1990, 8, 17, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	saved, _ := users.ByID(ctx, u.ID)
	if saved.Birthday != "1990-08-17" || saved.Zodiac != "Leo" {
		t.Errorf("saved = %+v", saved)
	}
}

func TestBirthChartPromoIsCut(t *testing.T) {
	promo := "Discover the key to your unique life path and personality with your premium Birth Chart."
	r, _, tr := newReadings(map[string]string{
		"https://h/us/tarot/tarot-true-love.aspx": `<div class="grid">Your Reading These are the cards of love.
` + promo + "\n" + promo + "\nTrue Love Tarot Reading</div>",
		"https://h/us/tarot/tarot-gems.aspx": `<div class="grid">Gems Oracle Wellness: Malachite
Courage.
` + promo + "\n" + promo + "\nToday's Tip:</div>",
	})
	res, err := r.TarotTrueLove(context.Background(), 3, 9, "ID")
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(res.Lines, "|"); strings.Contains(got, "Discover") || strings.Contains(got, "True Love Tarot Reading") {
		t.Errorf("promo must be cut before translating: %q", got)
	}
	if tr.calls != 1 {
		t.Errorf("translate calls = %d", tr.calls)
	}
	res, err = r.Clairvoyance(context.Background(), 4, "Viky", "EN")
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(CleanLines(res.Lines), "|"); got != "Wellness: Malachite|Courage." {
		t.Errorf("clairvoyance lines = %q", got)
	}
}

type failingTranslator struct{}

func (failingTranslator) Translate(context.Context, string, string, string) ([]string, error) {
	return nil, errors.New("upstream returned 429")
}

func TestTranslationFallbackShowsOriginalWithNotice(t *testing.T) {
	r, _, _ := newReadings(map[string]string{
		"https://h/us/horoscopes/general/horoscope-general-daily-today.aspx": `<div class="main-horoscope"><p>Oct 1, 2026 - Bold moves pay off.</p></div>`,
		"https://p/arti_nama.php": `<div id="body">ARTI NAMA` + ads + `
Nama Spirit, memiliki arti: Analitis.
Nama: </div>`,
	})
	r.Translate = failingTranslator{}
	in := HoroscopeInput{Birthday: time.Date(1998, 10, 1, 0, 0, 0, 0, time.UTC)}

	res, err := r.Horoscope(context.Background(), 2, in, "ID")
	if err != nil {
		t.Fatalf("a failed translation must not fail the reading: %v", err)
	}
	if strings.Join(res.Lines, "|") != "Bold moves pay off." || !strings.Contains(res.Notice, "Terjemahan sedang tidak tersedia") {
		t.Errorf("ID fallback = %q / %q", res.Lines, res.Notice)
	}

	res, err = r.Primbon(context.Background(), 1, PrimbonInput{Name: "Spirit Walker"}, "EN")
	if err != nil || !strings.Contains(strings.Join(res.Lines, " "), "Analitis") || !strings.Contains(res.Notice, "original language") {
		t.Errorf("EN fallback = %q / %q / %v", res.Lines, res.Notice, err)
	}

	// No notice when nothing had to be translated.
	res, _ = r.Horoscope(context.Background(), 2, in, "EN")
	if res.Notice != "" {
		t.Errorf("unexpected notice %q", res.Notice)
	}
}
