package main

// Run go run main.go -term 25F -save

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file from parent directory
	_ = godotenv.Load("../.env")
	var term string
	var save bool
	flag.StringVar(&term, "term", "25F", "UCLA term code (e.g., 25F, 25S)")
	flag.BoolVar(&save, "save", false, "Upsert results into Postgres")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	raw, err := FetchSubjectAreas(ctx, term)
	if err != nil {
		log.Fatalf("fetch failed: %v", err)
	}

	areas, err := ParseSubjectAreas(raw)
	if err != nil {
		log.Fatalf("parse failed: %v", err)
	}

	for i, a := range areas {
		fmt.Printf("%d\t%s\t%s\n", i+1, a.Name, a.Code)
	}

	if save {
		if err := SaveSubjectAreas(ctx, areas); err != nil {
			log.Fatalf("save failed: %v", err)
		}
		fmt.Printf("Saved %d subject areas to Postgres.\n", len(areas))
	}
}
