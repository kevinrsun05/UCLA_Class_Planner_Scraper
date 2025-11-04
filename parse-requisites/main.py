import os
import json
import psycopg2
from dotenv import load_dotenv
from parser import parse_requisites   # from parser.py above

# --------------------------------------
# Environment
# --------------------------------------
load_dotenv()
DB_URL = os.getenv("DATABASE_URL")


# --------------------------------------
# Utilities
# --------------------------------------
def load_subjects(conn):
    """Load subject names & codes from DB once, for fast in-memory lookup."""
    with conn.cursor() as cur:
        cur.execute("SELECT id, code, name FROM subject_areas;")
        rows = cur.fetchall()

    # Build lowercase sets for quick lookup
    names = {r[2].lower() for r in rows}
    codes = {r[1].lower(): r[2] for r in rows}   # map code → name
    id_to_name = {r[0]: r[2] for r in rows}      # map id → canonical name
    return names, codes, id_to_name


def get_subject_hint(subject_id, id_to_name):
    """Map subject_area_id from courses table to canonical subject name."""
    return id_to_name.get(subject_id, None)


# --------------------------------------
# Main processor
# --------------------------------------
def process_requisites(limit_range=None):
    """
    limit_range: tuple (start_id, end_id) to restrict processing,
    e.g. (205, 225)
    """
    conn = psycopg2.connect(DB_URL)
    valid_subjects, subject_codes, id_to_name = load_subjects(conn)

    with conn.cursor() as cur:
        base_query = """
            SELECT id, requisites_text, subject_area_id
            FROM courses
            WHERE requisites_text IS NOT NULL
            AND requisites_parsed IS NULL
        """
        if limit_range:
            base_query += f" AND id BETWEEN {limit_range[0]} AND {limit_range[1]}"
        base_query += ";"
        cur.execute(base_query)
        rows = cur.fetchall()

        print(f"🔍 Found {len(rows)} courses to process")

        updated_count = 0
        skipped_count = 0

        for course_id, text, subject_id in rows:
            subj_hint = get_subject_hint(subject_id, id_to_name)
            if not subj_hint or not text:
                continue

            parsed = parse_requisites(text, subj_hint, valid_subjects)

            # ❗ Skip if parser returned None (non-basic case)
            if parsed is None:
                skipped_count += 1
                continue

            cur.execute("""
                UPDATE courses
                SET requisites_parsed = %s
                WHERE id = %s;
            """, (json.dumps(parsed), course_id))
            updated_count += 1
            print(f"✅ Updated course {course_id}: {subj_hint}")

    conn.commit()
    conn.close()
    print(f"🎉 Done updating requisites. Updated: {updated_count}, Skipped (advanced): {skipped_count}")


# --------------------------------------
# Entry point
# --------------------------------------
if __name__ == "__main__":
    # Example: process a small range for testing
    #process_requisites((205, 225))

    # Or uncomment to process all
    process_requisites()
