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
	var term string
	var save bool
	var all bool
	var subject string

	flag.StringVar(&term, "term", "25F", "UCLA term code (e.g., 25F, 25S)")
	flag.BoolVar(&save, "save", false, "Save course offerings to database")
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
			courses, err := FetchAndParseCourses(ctx, sa, term)
			if err != nil {
				log.Printf("skip %s: %v", sa.Code, err)
				continue
			}
			for _, c := range courses {
				fmt.Printf("%s\t%s\n", c.Number, c.Title)
			}
			if save {
				if err := SaveCourseOfferings(ctx, db, term, courses); err != nil {
					log.Printf("save %s failed: %v", sa.Code, err)
					continue
				}
			}
			total += len(courses)
		}
		fmt.Printf("\nDone. Parsed %d course offerings for term %s across %d subject areas.\n", total, term, len(subjects))

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

			courses, err := FetchAndParseCourses(ctx, sa, term)
			if err != nil {
				log.Fatalf("fetch/parse failed: %v", err)
			}
			for _, c := range courses {
				fmt.Printf("%s\t%s\n", c.Number, c.Title)
			}
			if err := SaveCourseOfferings(ctx, db, term, courses); err != nil {
				log.Fatalf("save failed: %v", err)
			}
			fmt.Printf("Saved %d course offerings for %s in term %s.\n", len(courses), subject, term)
		} else {
			// print-only: use a fake ID since we won't save
			sa := SubjectArea{ID: 0, Code: subject, Name: subject}
			courses, err := FetchAndParseCourses(ctx, sa, term)
			if err != nil {
				log.Fatalf("fetch/parse failed: %v", err)
			}
			for _, c := range courses {
				fmt.Printf("%s\t%s\n", c.Number, c.Title)
			}
			fmt.Printf("Parsed %d courses for %s.\n", len(courses), subject)
		}

	default:
		log.Fatalf("usage:\n  go run . --term 25F --subject \"COM SCI\" [--save]\n  go run . --term 25F --all [--save]")
	}
}
