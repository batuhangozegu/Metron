package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct{ Pool *pgxpool.Pool }

const userColumns = `id::text, email, totp_enabled, created_at, password_hash, totp_secret_enc, totp_last_step`

func scanUser(row pgx.Row) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.TwoFactorEnabled, &u.CreatedAt, &u.PasswordHash, &u.Secret, &u.LastStep)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrUnauthorized
	}
	u.CreatedAt = u.CreatedAt.UTC()
	return u, err
}

func (p Postgres) Register(ctx context.Context, email, hash string, session Session) (User, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback(ctx)
	user, err := scanUser(tx.QueryRow(ctx, `INSERT INTO app.users(email,password_hash) VALUES($1,$2) RETURNING `+userColumns, email, hash))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return User{}, ErrConflict
		}
		return User{}, err
	}
	session.UserID = user.ID
	if err = insertSession(ctx, tx, session); err != nil {
		return User{}, err
	}
	return user, tx.Commit(ctx)
}

func (p Postgres) UserByEmail(ctx context.Context, email string) (User, error) {
	return scanUser(p.Pool.QueryRow(ctx, `SELECT `+userColumns+` FROM app.users WHERE lower(email)=$1`, email))
}

func (p Postgres) UserByID(ctx context.Context, id string) (User, error) {
	return scanUser(p.Pool.QueryRow(ctx, `SELECT `+userColumns+` FROM app.users WHERE id=$1`, id))
}

func insertSession(ctx context.Context, tx pgx.Tx, s Session) error {
	_, err := tx.Exec(ctx, `INSERT INTO app.sessions(user_id,token_hash,pending_2fa,created_at,expires_at,last_seen_at,user_agent,ip)
 VALUES($1,$2,$3,$4,$5,$4,$6,$7)`, s.UserID, s.Hash, s.Pending, s.CreatedAt, s.ExpiresAt, s.UserAgent, s.IP)
	return err
}

func (p Postgres) InsertSession(ctx context.Context, s Session) error {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// Serialize with 2FA state changes so password-only logins cannot race enable.
	var enabled bool
	if err = tx.QueryRow(ctx, `SELECT totp_enabled FROM app.users WHERE id=$1 FOR UPDATE`, s.UserID).Scan(&enabled); err != nil {
		return err
	}
	if enabled != s.Pending {
		return ErrState
	}
	if _, err = tx.Exec(ctx, `DELETE FROM app.sessions WHERE user_id=$1 AND expires_at <= $2`, s.UserID, s.CreatedAt); err != nil {
		return err
	}
	if err = insertSession(ctx, tx, s); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p Postgres) ResolveSession(ctx context.Context, hash []byte, pending bool, now time.Time) (User, error) {
	// CTE atomically tests expiry before touching it; pending sessions never slide.
	return scanUser(p.Pool.QueryRow(ctx, `WITH valid AS (
 UPDATE app.sessions SET last_seen_at=$3,
 expires_at=CASE WHEN pending_2fa THEN expires_at ELSE $4 END
 WHERE token_hash=$1 AND pending_2fa=$2 AND expires_at>$3 RETURNING user_id)
 SELECT `+userColumns+` FROM app.users JOIN valid ON users.id=valid.user_id`, hash, pending, now, now.Add(SessionTTL)))
}

func (p Postgres) DeleteSession(ctx context.Context, hash []byte) error {
	_, err := p.Pool.Exec(ctx, `DELETE FROM app.sessions WHERE token_hash=$1`, hash)
	return err
}

func (p Postgres) SetSecret(ctx context.Context, id string, secret []byte, now time.Time) error {
	result, err := p.Pool.Exec(ctx, `UPDATE app.users SET totp_secret_enc=$2,totp_last_step=NULL,updated_at=$3 WHERE id=$1 AND NOT totp_enabled`, id, secret, now)
	if err == nil && result.RowsAffected() != 1 {
		return ErrState
	}
	return err
}

func (p Postgres) ApplyTOTP(ctx context.Context, c TOTPChange) error {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	wasEnabled := c.Action != "enable"
	enabled := c.Action != "disable"
	result, err := tx.Exec(ctx, `UPDATE app.users SET totp_enabled=$5,
 totp_secret_enc=CASE WHEN $5 THEN totp_secret_enc ELSE NULL END,
 totp_last_step=$3,updated_at=$4
 WHERE id=$1 AND totp_secret_enc=$2 AND (totp_last_step IS NULL OR totp_last_step<$3) AND totp_enabled=$6`, c.UserID, c.Secret, c.Step, c.Now, enabled, wasEnabled)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return ErrCode
	}
	if c.Action == "login" {
		result, err = tx.Exec(ctx, `DELETE FROM app.sessions WHERE token_hash=$1 AND user_id=$2 AND pending_2fa AND expires_at>$3`, c.CurrentHash, c.UserID, c.Now)
		if err != nil {
			return err
		}
		if result.RowsAffected() != 1 {
			return ErrUnauthorized
		}
		if err = insertSession(ctx, tx, c.NewSession); err != nil {
			return err
		}
	} else {
		// Keep only the authenticated caller. This also invalidates pending logins.
		if _, err = tx.Exec(ctx, `DELETE FROM app.sessions WHERE user_id=$1 AND token_hash<>$2`, c.UserID, c.CurrentHash); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
