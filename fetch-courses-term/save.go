package main

// go run . --term 25F --all --save
// 26S, 26W, 25F, 25W, 25S, 24F

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const upsertCourse = `
INSERT INTO public.courses (subject_area_id, number, title, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
ON CONFLICT (subject_area_id, number) DO UPDATE
SET title = EXCLUDED.title,
    updated_at = NOW()
RETURNING id;
`

const upsertOffering = `
INSERT INTO public.course_offerings (course_id, term, section, created_at, updated_at)
VALUES ($1, $2, NULL, NOW(), NOW())
ON CONFLICT (course_id, term, section) DO UPDATE
SET updated_at = NOW();
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

func SaveCourseOfferings(ctx context.Context, db *sql.DB, term string, courses []Course) error {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	courseStmt, err := tx.PrepareContext(ctx, upsertCourse)
	if err != nil {
		return fmt.Errorf("prepare course: %w", err)
	}
	defer courseStmt.Close()

	offeringStmt, err := tx.PrepareContext(ctx, upsertOffering)
	if err != nil {
		return fmt.Errorf("prepare offering: %w", err)
	}
	defer offeringStmt.Close()

	for _, c := range courses {
		// First, ensure course exists in courses table
		var courseID int64
		err := courseStmt.QueryRowContext(ctx, c.SubjectAreaID, c.Number, c.Title).Scan(&courseID)
		if err != nil {
			return fmt.Errorf("upsert course %q: %w", c.Number, err)
		}

		// Then, create offering entry for this term
		_, err = offeringStmt.ExecContext(ctx, courseID, term)
		if err != nil {
			return fmt.Errorf("upsert offering %q %s: %w", c.Number, term, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// Helpers to read subject areas from DB
func GetAllSubjectAreas(ctx context.Context, db *sql.DB) ([]SubjectArea, error) {
	rows, err := db.QueryContext(ctx, `select id, name, code from public.subject_areas order by code`)
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

func GetSubjectAreaByCode(ctx context.Context, db *sql.DB, code string) (SubjectArea, error) {
	var sa SubjectArea
	err := db.QueryRowContext(ctx, `select id, name, code from public.subject_areas where code=$1`, code).Scan(&sa.ID, &sa.Name, &sa.Code)
	return sa, err
}
