package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver (needs cgo / a C compiler)

	"github.com/xoxyn/semul/internal/domain"
)

const schema = `
CREATE TABLE IF NOT EXISTS contacted (
	company_key TEXT PRIMARY KEY,
	email_key   TEXT NOT NULL,
	company     TEXT NOT NULL,
	email       TEXT NOT NULL,
	subject     TEXT NOT NULL,
	sent_at     TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_contacted_email ON contacted(email_key);
`

// SQLite remembers which prospects have been emailed. It implements dedupe.Store.
type SQLite struct {
	db *sql.DB
}

// Open opens (and creates, if needed) the database file at path.
func Open(path string) (*SQLite, error) {
	db, err := sql.Open("sqlite3", "file:"+path+"?_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	db.SetMaxOpenConns(1) // SQLite is happiest with a single writer

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}
	return &SQLite{db: db}, nil
}

func (s *SQLite) Close() error { return s.db.Close() }

// Contacted reports whether the prospect's company or email was already emailed.
func (s *SQLite) Contacted(ctx context.Context, p domain.Prospect) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM contacted WHERE company_key = ? OR email_key = ?)`,
		p.CompanyKey(), p.EmailKey(),
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("query contacted: %w", err)
	}
	return exists, nil
}

// MarkSent records a successful send. Recording the same company twice is a no-op.
func (s *SQLite) MarkSent(ctx context.Context, job domain.EmailJob) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO contacted
			(company_key, email_key, company, email, subject, sent_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		job.CompanyKey(), job.EmailKey(), job.Company, job.Email,
		job.Draft.Subject, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("insert contacted: %w", err)
	}
	return nil
}

