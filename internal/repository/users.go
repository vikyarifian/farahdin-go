package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/vikyarifian/farahdin-go/internal/domain"
)

// Users is the users table.
type Users struct {
	DB *sql.DB
}

const userColumns = `id, username, fullname, email, COALESCE(birthday, ''), COALESCE(birthplace, ''),
	COALESCE(gender, ''), COALESCE(zodiac, ''), image_url, COALESCE(google_sub, ''), COALESCE(clerk_id, '')`

func scanUser(row interface{ Scan(...any) error }) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Username, &u.Fullname, &u.Email, &u.Birthday, &u.Birthplace,
		&u.Gender, &u.Zodiac, &u.ImageURL, &u.GoogleSub, &u.ClerkID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &u, nil
}

// ByID returns the user with id.
func (r *Users) ByID(ctx context.Context, id int64) (*domain.User, error) {
	return scanUser(r.DB.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id))
}

// ByGoogleSub ports the by_clerk_id lookup for the new identity provider.
func (r *Users) ByGoogleSub(ctx context.Context, sub string) (*domain.User, error) {
	return scanUser(r.DB.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE google_sub = $1`, sub))
}

// FirstByEmail ports the by_email index lookup with .first().
func (r *Users) FirstByEmail(ctx context.Context, email string) (*domain.User, error) {
	return scanUser(r.DB.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE email = $1 ORDER BY id LIMIT 1`, email))
}

// Insert creates a user and returns its id.
func (r *Users) Insert(ctx context.Context, u *domain.User) (int64, error) {
	var id int64
	err := r.DB.QueryRowContext(ctx, `INSERT INTO users
		(username, fullname, email, birthday, birthplace, gender, zodiac, google_sub, image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		u.Username, u.Fullname, u.Email, u.Birthday, u.Birthplace, u.Gender, u.Zodiac, nullable(u.GoogleSub), u.ImageURL).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}
	return id, nil
}

// LinkIdentity ports the createUser patch of an existing email match:
// username, fullname and the provider id are overwritten.
func (r *Users) LinkIdentity(ctx context.Context, id int64, username, fullname, sub, imageURL string) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE users SET username = $1, fullname = $2, google_sub = $3, image_url = $4,
		updated_at = now() WHERE id = $5`,
		username, fullname, sub, imageURL, id)
	if err != nil {
		return fmt.Errorf("link identity: %w", err)
	}
	return nil
}

// SetImage stores the provider's profile photo URL.
func (r *Users) SetImage(ctx context.Context, id int64, imageURL string) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE users SET image_url = $1 WHERE id = $2 AND image_url <> $1`, imageURL, id)
	return err
}

// UpdateProfile ports users.updateUser for the fields the user may edit.
func (r *Users) UpdateProfile(ctx context.Context, id int64, fullname, birthday, birthplace, gender, zodiac string) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE users SET fullname = $1, birthday = $2, birthplace = $3, gender = $4, zodiac = $5,
		updated_at = now() WHERE id = $6`,
		fullname, birthday, birthplace, gender, zodiac, id)
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	return nil
}

// UpdateBirthday stores a new birthday and its zodiac sign.
func (r *Users) UpdateBirthday(ctx context.Context, id int64, birthday, zodiac string) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE users SET birthday = $1, zodiac = $2,
		updated_at = now() WHERE id = $3`, birthday, zodiac, id)
	if err != nil {
		return fmt.Errorf("update birthday: %w", err)
	}
	return nil
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
