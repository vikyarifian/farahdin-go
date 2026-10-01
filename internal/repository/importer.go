package repository

import (
	"archive/zip"
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
)

// ImportStats counts imported rows per table.
type ImportStats struct {
	Users, Categories, Inboxes int
}

// ImportConvex loads a `npx convex export` ZIP (one <table>/documents.jsonl
// per table) into the database. Rows are keyed by their Convex _id, so
// running the import again updates instead of duplicating.
func ImportConvex(ctx context.Context, db *sql.DB, zipPath string) (ImportStats, error) {
	var st ImportStats
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return st, fmt.Errorf("open export: %w", err)
	}
	defer zr.Close()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return st, err
	}
	defer tx.Rollback()

	err = eachDoc(&zr.Reader, "users", func(raw json.RawMessage) error {
		var d struct {
			ID         string  `json:"_id"`
			Username   string  `json:"username"`
			Fullname   string  `json:"fullname"`
			Email      string  `json:"email"`
			Birthday   *string `json:"birthday"`
			Birthplace *string `json:"birthplace"`
			Gender     *string `json:"gender"`
			Zodiac     *string `json:"zodiac"`
			ClerkID    string  `json:"clerkId"`
		}
		if err := json.Unmarshal(raw, &d); err != nil {
			return err
		}
		if d.ID == "" || d.Email == "" {
			return errors.New("user without _id or email")
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO users
			(convex_id, username, fullname, email, birthday, birthplace, gender, zodiac, clerk_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (convex_id) DO UPDATE SET username = excluded.username, fullname = excluded.fullname,
				email = excluded.email, birthday = excluded.birthday, birthplace = excluded.birthplace,
				gender = excluded.gender, zodiac = excluded.zodiac, clerk_id = excluded.clerk_id`,
			d.ID, d.Username, d.Fullname, d.Email, d.Birthday, d.Birthplace, d.Gender, d.Zodiac, d.ClerkID)
		st.Users++
		return err
	})
	if err != nil {
		return st, fmt.Errorf("users: %w", err)
	}

	err = eachDoc(&zr.Reader, "categories", func(raw json.RawMessage) error {
		var d struct {
			ID     string `json:"_id"`
			Code   string `json:"code"`
			NameID string `json:"nameID"`
			NameEN string `json:"nameEN"`
		}
		if err := json.Unmarshal(raw, &d); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO categories (convex_id, code, name_id, name_en) VALUES ($1, $2, $3, $4)
			ON CONFLICT (convex_id) DO UPDATE SET code = excluded.code, name_id = excluded.name_id, name_en = excluded.name_en`,
			d.ID, d.Code, d.NameID, d.NameEN)
		st.Categories++
		return err
	})
	if err != nil {
		return st, fmt.Errorf("categories: %w", err)
	}

	err = eachDoc(&zr.Reader, "inboxes", func(raw json.RawMessage) error {
		var d struct {
			ID         string `json:"_id"`
			UserID     string `json:"userId"`
			CategoryID string `json:"categoryId"`
			MessageEN  string `json:"messageEN"`
			MessageID  string `json:"messageID"`
			Date       string `json:"date"`
		}
		if err := json.Unmarshal(raw, &d); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO inboxes (convex_id, user_id, category_id, message_en, message_id, date)
			VALUES ($1, (SELECT id FROM users WHERE convex_id = $2), (SELECT id FROM categories WHERE convex_id = $3), $4, $5, $6)
			ON CONFLICT (convex_id) DO UPDATE SET message_en = excluded.message_en, message_id = excluded.message_id, date = excluded.date`,
			d.ID, d.UserID, d.CategoryID, d.MessageEN, d.MessageID, d.Date)
		st.Inboxes++
		return err
	})
	if err != nil {
		return st, fmt.Errorf("inboxes: %w", err)
	}
	return st, tx.Commit()
}

// eachDoc calls fn for every line of <table>/documents.jsonl; a missing
// table is not an error (Convex omits empty tables).
func eachDoc(zr *zip.Reader, table string, fn func(json.RawMessage) error) error {
	f, err := zr.Open(path.Join(table, "documents.jsonl"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64<<10), 16<<20)
	line := 0
	for sc.Scan() {
		line++
		if len(sc.Bytes()) == 0 {
			continue
		}
		if err := fn(json.RawMessage(sc.Bytes())); err != nil {
			return fmt.Errorf("line %d: %w", line, err)
		}
	}
	if err := sc.Err(); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
