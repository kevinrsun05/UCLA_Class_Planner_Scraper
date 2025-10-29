# UCLA Class Planner Scraper

A collection of Go-based scrapers to fetch UCLA course data including subject areas, all catalog courses, and term-specific course offerings.

## Setup

### 1. Install Dependencies

Each scraper has its own Go module. Install dependencies for the scrapers you plan to use:

```bash
# Install dependencies for all scrapers
cd fetch-subject-areas && go mod download && cd ..
cd fetch-all-courses && go mod download && cd ..
cd fetch-courses-term && go mod download && cd ..
```

### 2. Configure Database Connection

Copy the example environment file and add your database credentials:

```bash
cp .env.example .env
```

Edit `.env` and set your DATABASE_URL:

```
DATABASE_URL=postgresql://user:password@host:5432/database?sslmode=require
```

**Note:** The `.env` file is automatically loaded by all scrapers. Never commit this file to version control.

## Available Term Codes

UCLA uses term codes in the format `{YY}{S}` where:
- `YY` = last two digits of the year
- `S` = season code: `W` (Winter), `S` (Spring), `F` (Fall)

Examples: `25F` (Fall 2025), `25W` (Winter 2025), `25S` (Spring 2025), `26F` (Fall 2026)

## Database Architecture

The scrapers populate three tables:

**1. `subject_areas`** - UCLA departments
- `id`, `code`, `name`
- Example: `{id: 5, code: "COM SCI", name: "Computer Science"}`

**2. `courses`** - All courses in UCLA's catalog (timeless data)
- `id`, `subject_area_id`, `number`, `title`, `description`, `units`, `requisites_text`
- Example: `{number: "111", title: "Operating Systems", units: "5.0"}`

**3. `course_offerings`** - Which courses are offered in which terms
- `id`, `course_id`, `term`, `section`, `instructor`, `meeting_times`, `location`, `enrollment_status`
- Example: `{course_id: 42, term: "25F", section: "Lec 1"}`
- **Note:** Section details (instructor, times, location) are not yet parsed

## Scrapers

### 1. Subject Areas Scraper

Fetches all subject areas (departments) available for a given term.

**Populates:** `subject_areas` table

**Run from:** `fetch-subject-areas/`

```bash
cd fetch-subject-areas

# Print subject areas only (no database save)
go run . --term 25F

# Fetch and save to database
go run . --term 25F --save
```

**Flags:**
- `--term` - UCLA term code (default: `25F`)
- `--save` - Save results to Postgres database

**Example output:**
```
1    Aerospace Engineering    AEROSP
2    African Studies          AF AMER
3    Computer Science         COM SCI
...
Saved 192 subject areas to Postgres.
```

---

### 2. All Courses Scraper (Catalog Data)

Fetches ALL courses from UCLA's catalog API, regardless of when they're offered.

**Populates:** `courses` table

**Run from:** `fetch-all-courses/`

```bash
cd fetch-all-courses

# Fetch courses for a single subject (print only, no save)
go run . --subject "COM SCI"

# Fetch and save courses for a single subject
go run . --subject "COM SCI" --save

# Fetch and save ALL catalog courses for ALL subject areas
go run . --all --save
```

**Flags:**
- `--subject` - Single subject area code (e.g., `"COM SCI"`, `"MATH"`)
- `--all` - Process all subject areas from database
- `--save` - Save results to Postgres database

**Example output:**
```
31      Introduction to Computer Science I      [4.0 units]
32      Introduction to Computer Science II     [4.0 units]
111     Operating Systems Principles            [5.0 units]
...
Saved 179 courses for COM SCI.
```

**Data captured:**
- Course number, title
- Full description
- Units (credit hours)
- Prerequisites/requisites text

---

### 3. Course Offerings Scraper (Term-Specific)

Fetches which courses are actually offered in a specific term.

**Populates:** `course_offerings` table

**Run from:** `fetch-courses-term/`

```bash
cd fetch-courses-term

# Fetch offerings for a single subject in a term
go run . --term 25F --subject "COM SCI" --save

# Fetch offerings for ALL subjects in a term
go run . --term 25F --all --save
```

**Flags:**
- `--term` - UCLA term code (e.g., `25F`, `26W`)
- `--subject` - Single subject area code (e.g., `"COM SCI"`)
- `--all` - Process all subject areas from database
- `--save` - Save results to Postgres database

**Example output:**
```
31      Introduction to Computer Science I
111     Operating Systems Principles
131     Programming Languages
...
Saved 25 course offerings for COM SCI in term 25F.
```

**Current limitation:** This scraper only tracks which courses are offered in which terms. Section details (instructor, meeting times, location, enrollment status) are **not yet parsed**. Full section parsing will be implemented in a future update.

---

## Typical Workflow

### Initial Setup: Populate Full Catalog

```bash
# 1. Fetch all subject areas
cd fetch-subject-areas
go run . --term 25F --save

# 2. Fetch ALL catalog courses (descriptions, units, prerequisites)
cd ../fetch-all-courses
go run . --all --save
```

### Update Term Offerings

When a new term's schedule is released, run:

```bash
# 3. Fetch which courses are offered this term
cd fetch-courses-term
go run . --term 26W --all --save
```

This workflow gives you:
- Complete catalog of all UCLA courses
- Which courses are offered in specific terms
- Can query: "What courses does COM SCI offer?" vs "What COM SCI courses are available in Fall 2025?"

## Future Enhancements

- [ ] Parse full section details in `fetch-courses-term` (instructor, meeting times, location, enrollment)
- [ ] Add cross-department prerequisite parsing
- [ ] Create automated scraping script for multiple terms
