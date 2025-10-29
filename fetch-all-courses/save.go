package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const upsertCourse = `
INSERT INTO public.courses (subject_area_id, number, title, description, units, requisites_text, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
ON CONFLICT (subject_area_id, number) DO UPDATE
SET title = EXCLUDED.title,
    description = EXCLUDED.description,
    units = EXCLUDED.units,
    requisites_text = EXCLUDED.requisites_text,
    updated_at = NOW();
`

func openDB(ctx context.Context) (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx2); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return db, nil
}

func UpsertCourses(ctx context.Context, db *sql.DB, subjectAreaID int64, courses []Course) error {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, upsertCourse)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	// Upsert course details (insert or update)
	for _, c := range courses {
		_, err := stmt.ExecContext(ctx, subjectAreaID, c.Number, c.Title, c.Description, c.Units, c.RequisitesText)
		if err != nil {
			return fmt.Errorf("upsert %q: %w", c.Number, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// Get all subject areas from DB
func GetAllSubjectAreas(ctx context.Context, db *sql.DB) ([]SubjectArea, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, name, code FROM public.subject_areas ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []SubjectArea
	for rows.Next() {
		var sa SubjectArea
		if err := rows.Scan(&sa.ID, &sa.Name, &sa.Code); err != nil {
			return nil, err
		}
		out = append(out, sa)
	}
	return out, rows.Err()
}

// Get subject area by code
func GetSubjectAreaByCode(ctx context.Context, db *sql.DB, code string) (SubjectArea, error) {
	var sa SubjectArea
	err := db.QueryRowContext(ctx, `SELECT id, name, code FROM public.subject_areas WHERE code=$1`, code).Scan(&sa.ID, &sa.Name, &sa.Code)
	return sa, err
}
