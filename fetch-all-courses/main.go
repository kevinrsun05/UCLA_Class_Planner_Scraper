package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env file from parent directory
	_ = godotenv.Load("../.env")

	var save bool
	var all bool
	var subject string

	flag.BoolVar(&save, "save", false, "Save courses to database")
	flag.BoolVar(&all, "all", false, "Process all subject areas from DB")
	flag.StringVar(&subject, "subject", "", "Single subject area code (e.g., COM SCI)")
	flag.Parse()

	ctx := context.Background()

	switch {
	case all:
		db, err := openDB(ctx)
		if err != nil {
			log.Fatalf("db open failed: %v", err)
		}
		defer db.Close()

		subjects, err := GetAllSubjectAreas(ctx, db)
		if err != nil {
			log.Fatalf("load subject areas: %v", err)
		}

		total := 0
		for _, sa := range subjects {
			fmt.Printf("\n== %s — %s ==\n", sa.Code, sa.Name)

			data, err := FetchCourseDescriptions(ctx, sa)
			if err != nil {
				log.Printf("skip %s: %v", sa.Code, err)
				continue
			}

			courses, err := ParseCourseDescriptions(sa, data)
			if err != nil {
				log.Printf("parse failed for %s: %v", sa.Code, err)
				continue
			}

			for _, c := range courses {
				fmt.Printf("%s\t%s\t[%s units]\n", c.Number, c.Title, c.Units)
			}

			if save {
				if err := UpsertCourses(ctx, db, sa.ID, courses); err != nil {
					log.Printf("save %s failed: %v", sa.Code, err)
					continue
				}
			}
			total += len(courses)
		}
		fmt.Printf("\nDone. Processed %d courses across %d subject areas.\n", total, len(subjects))

	case subject != "":
		if save {
			db, err := openDB(ctx)
			if err != nil {
				log.Fatalf("db open failed: %v", err)
			}
			defer db.Close()

			sa, err := GetSubjectAreaByCode(ctx, db, subject)
			if err != nil {
				log.Fatalf("subject area %q not found in DB (needed for --save): %v", subject, err)
			}

			data, err := FetchCourseDescriptions(ctx, sa)
			if err != nil {
				log.Fatalf("fetch failed: %v", err)
			}

			courses, err := ParseCourseDescriptions(sa, data)
			if err != nil {
				log.Fatalf("parse failed: %v", err)
			}

			for _, c := range courses {
				fmt.Printf("%s\t%s\t[%s units]\n", c.Number, c.Title, c.Units)
			}

			if err := UpsertCourses(ctx, db, sa.ID, courses); err != nil {
				log.Fatalf("save failed: %v", err)
			}
			fmt.Printf("Saved %d courses for %s.\n", len(courses), subject)
		} else {
			// print-only: use a fake ID since we won't save
			sa := SubjectArea{ID: 0, Code: subject, Name: subject}
			data, err := FetchCourseDescriptions(ctx, sa)
			if err != nil {
				log.Fatalf("fetch failed: %v", err)
			}

			courses, err := ParseCourseDescriptions(sa, data)
			if err != nil {
				log.Fatalf("parse failed: %v", err)
			}

			for _, c := range courses {
				fmt.Printf("%s\t%s\t[%s units]\n", c.Number, c.Title, c.Units)
			}
			fmt.Printf("Fetched %d courses for %s.\n", len(courses), subject)
		}

	default:
		log.Fatalf("usage:\n  go run . --subject \"COM SCI\" [--save]\n  go run . --all [--save]")
	}
}
