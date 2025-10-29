// save.go
package main

// go run . --term 25F --save
// 26S, 26W, 25F, 25W, 25S, 24F

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const upsertSubjectArea = `
INSERT INTO public.subject_areas (name, code, created_at, updated_at)
VALUES ($1, $2, NOW(), NOW())
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    updated_at = NOW();
`

func SaveSubjectAreas(ctx context.Context, areas []SubjectArea) error {
	dsn := os.Getenv("DATABASE_URL") // e.g. postgres://postgres:PASS@db.PROJ.supabase.co:5432/postgres?sslmode=require
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer db.Close()

	// Optional: tighten timeouts
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	stmt, err := tx.PrepareContext(ctx, upsertSubjectArea)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	for _, a := range areas {
		if _, err := stmt.ExecContext(ctx, a.Name, a.Code); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("upsert %q: %w", a.Code, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
