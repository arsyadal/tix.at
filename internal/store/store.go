package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ DB *pgxpool.Pool }

type ExpiredBooking struct {
	ID      string
	EventID string
}

func New(ctx context.Context, url string) (*Store, error) {
	db, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec(ctx, `alter table bookings add column if not exists expires_at timestamptz`); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec(ctx, `create index if not exists bookings_expires_at_idx on bookings(expires_at)`); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec(ctx, `create unique index if not exists bookings_user_event_active_idx on bookings(user_id, event_id) where status != 'CANCELLED'`); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{DB: db}, nil
}

func (s *Store) Close() { s.DB.Close() }

func (s *Store) UpsertEvent(ctx context.Context, id, name string, stock int) error {
	_, err := s.DB.Exec(ctx, `
		insert into events (id, name, stock) values ($1, $2, $3)
		on conflict (id) do update set name = excluded.name, stock = excluded.stock`, id, name, stock)
	return err
}

func (s *Store) CreateBooking(ctx context.Context, id, eventID, userID string, ttl time.Duration) error {
	_, err := s.DB.Exec(ctx, `insert into bookings (id, event_id, user_id, status, expires_at) values ($1, $2, $3, 'PENDING', now() + $4::interval)`, id, eventID, userID, ttl.String())
	return err
}

func (s *Store) IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (s *Store) CancelBooking(ctx context.Context, bookingID string) error {
	_, err := s.DB.Exec(ctx, `update bookings set status = 'CANCELLED', updated_at = now() where id = $1 and status = 'PENDING'`, bookingID)
	return err
}

func (s *Store) ExpiredPending(ctx context.Context, limit int) ([]ExpiredBooking, error) {
	rows, err := s.DB.Query(ctx, `select id, event_id from bookings where status = 'PENDING' and expires_at <= now() order by expires_at limit $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ExpiredBooking
	for rows.Next() {
		var b ExpiredBooking
		if err := rows.Scan(&b.ID, &b.EventID); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) CancelExpired(ctx context.Context, bookingID string) (bool, error) {
	tag, err := s.DB.Exec(ctx, `update bookings set status = 'CANCELLED', updated_at = now() where id = $1 and status = 'PENDING' and expires_at <= now()`, bookingID)
	return tag.RowsAffected() == 1, err
}

func (s *Store) MarkPaid(ctx context.Context, bookingID, eventID string) (bool, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `update bookings set status = 'PAID', updated_at = now() where id = $1 and status = 'PENDING' and expires_at > now()`, bookingID)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() == 1 {
		if _, err := tx.Exec(ctx, `update events set stock = stock - 1 where id = $1`, eventID); err != nil {
			return false, err
		}
		return true, tx.Commit(ctx)
	}

	var status string
	err = tx.QueryRow(ctx, `select status from bookings where id = $1`, bookingID).Scan(&status)
	if err == pgx.ErrNoRows {
		return false, tx.Commit(ctx)
	}
	if err != nil {
		return false, err
	}
	return false, tx.Commit(ctx)
}
