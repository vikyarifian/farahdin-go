// Package repotest gives tests an isolated, migrated PostgreSQL schema.
//
// Set TEST_POSTGRES_URL to a database the test user may create schemas in
// (never the production database). Each test gets its own schema, dropped
// when the test ends; without TEST_POSTGRES_URL database tests are skipped.
package repotest

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/url"
	"os"
	"testing"

	"github.com/vikyarifian/farahdin-go/internal/repository"
)

// Open returns a database handle whose search_path is a fresh schema with
// all migrations applied.
func Open(t testing.TB) *sql.DB {
	t.Helper()
	base := os.Getenv("TEST_POSTGRES_URL")
	if base == "" {
		t.Skip("TEST_POSTGRES_URL not set; skipping database test")
	}
	b := make([]byte, 6)
	rand.Read(b)
	schema := "farahdin_test_" + hex.EncodeToString(b)

	admin, err := repository.Open(base)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec(`CREATE SCHEMA ` + schema); err != nil {
		admin.Close()
		t.Fatalf("create test schema (the TEST_POSTGRES_URL user needs CREATE on the database): %v", err)
	}
	t.Cleanup(func() {
		admin.Exec(`DROP SCHEMA ` + schema + ` CASCADE`)
		admin.Close()
	})

	u, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	db, err := repository.Open(u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := repository.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	return db
}
