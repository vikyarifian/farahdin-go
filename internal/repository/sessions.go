package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Sessions is the sessions table.
type Sessions struct {
	DB *sql.DB
}

// Create stores a session keyed by the hash of its token.
func (r *Sessions) Create(ctx context.Context, tokenHash string, userID int64, now, expires time.Time) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO sessions (token_hash, user_id, created_at, expires_at) VALUES ($1, $2, $3, $4)`,
		tokenHash, userID, now, expires)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// UserID returns the user of a session that has not expired at now.
func (r *Sessions) UserID(ctx context.Context, tokenHash string, now time.Time) (int64, error) {
	var id int64
	err := r.DB.QueryRowContext(ctx, `SELECT user_id FROM sessions WHERE token_hash = $1 AND expires_at > $2`,
		tokenHash, now).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("load session: %w", err)
	}
	return id, nil
}

// Delete removes one session.
func (r *Sessions) Delete(ctx context.Context, tokenHash string) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash)
	return err
}

// DeleteExpired removes sessions that expired before now.
func (r *Sessions) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	res, err := r.DB.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at <= $1`, now)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
