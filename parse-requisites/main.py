import os
import json
import psycopg2
from dotenv import load_dotenv
from parser import parse_requisites

load_dotenv()  # loads DATABASE_URL from .env
DB_URL = os.getenv("DATABASE_URL")

def process_requisites():
    conn = psycopg2.connect(DB_URL)
    cur = conn.cursor()

    # Get subject area mapping (for inferring subject codes)
    cur.execute("SELECT id, code FROM subject_areas;")
    subject_lookup = {row[0]: row[1] for row in cur.fetchall()}

    cur.execute("""
        SELECT id, subject_area_id, requisites_text
        FROM courses
        WHERE requisites_text IS NOT NULL;
    """)

    courses = cur.fetchall()
    print(f"Processing {len(courses)} courses...")

    for course_id, subject_area_id, text in courses:
        subject_hint = subject_lookup.get(subject_area_id, "")
        parsed = parse_requisites(text, subject_hint)

        cur.execute("""
            UPDATE courses
            SET requisites_parsed = %s
            WHERE id = %s;
        """, (json.dumps(parsed), course_id))

    conn.commit()
    print("✅ Updated requisites_parsed for all courses.")
    cur.close()
    conn.close()

if __name__ == "__main__":
    process_requisites()
