package repository_test

import (
	"archive/zip"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vikyarifian/farahdin-go/internal/domain"
	"github.com/vikyarifian/farahdin-go/internal/repository"
	"github.com/vikyarifian/farahdin-go/internal/repository/repotest"
)

func openTest(t *testing.T) (*repository.Users, *repository.Sessions) {
	t.Helper()
	db := repotest.Open(t)
	// Migrations are idempotent.
	if err := repository.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	return &repository.Users{DB: db}, &repository.Sessions{DB: db}
}

func TestUsers(t *testing.T) {
	users, _ := openTest(t)
	ctx := context.Background()
	id, err := users.Insert(ctx, &domain.User{Username: "viky", Fullname: "Viky", Email: "v@example.com", Birthday: "2026-10-01", GoogleSub: "g1"})
	if err != nil {
		t.Fatal(err)
	}
	u, err := users.ByGoogleSub(ctx, "g1")
	if err != nil || u.ID != id || u.Birthday != "2026-10-01" {
		t.Fatalf("ByGoogleSub = %+v, %v", u, err)
	}
	if _, err := users.ByGoogleSub(ctx, "missing"); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("missing sub: %v", err)
	}
	if err := users.UpdateProfile(ctx, id, "Viky A", "1990-08-17", "Jakarta", "Male", "Leo"); err != nil {
		t.Fatal(err)
	}
	if err := users.LinkIdentity(ctx, id, "vk", "Viky B", "g2", "https://img"); err != nil {
		t.Fatal(err)
	}
	u, _ = users.FirstByEmail(ctx, "v@example.com")
	if u.Fullname != "Viky B" || u.Zodiac != "Leo" || u.GoogleSub != "g2" || u.ImageURL != "https://img" || u.Birthplace != "Jakarta" {
		t.Errorf("user = %+v", u)
	}
	// Users without a subject (e.g. imported) can coexist: NULL is not unique.
	if _, err := users.Insert(ctx, &domain.User{Username: "a", Fullname: "A", Email: "a@example.com"}); err != nil {
		t.Fatal(err)
	}
	if _, err := users.Insert(ctx, &domain.User{Username: "b", Fullname: "B", Email: "b@example.com"}); err != nil {
		t.Fatal(err)
	}
}

func TestSessions(t *testing.T) {
	users, sessions := openTest(t)
	ctx := context.Background()
	id, _ := users.Insert(ctx, &domain.User{Username: "v", Fullname: "V", Email: "v@example.com"})
	now := time.Now()
	if err := sessions.Create(ctx, "hash", id, now, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if got, err := sessions.UserID(ctx, "hash", now); err != nil || got != id {
		t.Errorf("UserID = %d, %v", got, err)
	}
	if _, err := sessions.UserID(ctx, "hash", now.Add(2*time.Hour)); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expired session must not resolve: %v", err)
	}
	if n, _ := sessions.DeleteExpired(ctx, now.Add(2*time.Hour)); n != 1 {
		t.Errorf("DeleteExpired = %d", n)
	}
}

func TestImportConvex(t *testing.T) {
	users, _ := openTest(t)
	ctx := context.Background()
	zipPath := filepath.Join(t.TempDir(), "export.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	write := func(name, body string) {
		w, _ := zw.Create(name)
		w.Write([]byte(body))
	}
	write("users/documents.jsonl",
		`{"_id":"u1","_creationTime":1,"username":"viky","fullname":"Viky","email":"v@example.com","birthday":"1990-08-17","zodiac":"Leo","gender":"","birthplace":"","clerkId":"user_abc"}`+"\n"+
			`{"_id":"u2","_creationTime":2,"username":"f","fullname":"F","email":"f@example.com","clerkId":"user_def"}`+"\n")
	write("categories/documents.jsonl", `{"_id":"c1","code":"GEN","nameID":"Umum","nameEN":"General"}`+"\n")
	write("inboxes/documents.jsonl", `{"_id":"i1","userId":"u1","categoryId":"c1","messageEN":"Hi","messageID":"Hai","date":"10/1/2026"}`+"\n")
	zw.Close()
	f.Close()

	for i := 0; i < 2; i++ { // idempotent
		st, err := repository.ImportConvex(ctx, users.DB, zipPath)
		if err != nil {
			t.Fatal(err)
		}
		if st.Users != 2 || st.Categories != 1 || st.Inboxes != 1 {
			t.Errorf("stats = %+v", st)
		}
	}
	u, err := users.FirstByEmail(ctx, "v@example.com")
	if err != nil || u.Zodiac != "Leo" || u.ClerkID != "user_abc" || u.GoogleSub != "" {
		t.Errorf("user = %+v, %v", u, err)
	}
	var n int
	users.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n)
	if n != 2 {
		t.Errorf("users = %d", n)
	}
	users.DB.QueryRow(`SELECT COUNT(*) FROM inboxes WHERE user_id = $1`, u.ID).Scan(&n)
	if n != 1 {
		t.Errorf("inbox not linked to user")
	}
}
