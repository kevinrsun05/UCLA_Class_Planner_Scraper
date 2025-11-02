import re
import json

# ------------------------------------------------------
# 1️⃣ Classifier — decide which parser to use
# ------------------------------------------------------
def classify_requisite(raw_text: str) -> str:
    """Classify a requisite string as 'basic' or 'advanced'."""
    text = raw_text.lower()

    # Keywords that indicate complex logic (handled later by advanced parser)
    if any(
        kw in text
        for kw in [
            "one course from",
            "either",
            "both",
            "must also",
            "may be taken concurrently",
            "and one of",
            "and one course from",
        ]
    ):
        return "advanced"

    # Simple course lists or single-course enforcement
    if re.search(r"\bcourses?\b", text) or re.search(r"\bor\b", text):
        return "basic"

    # Default to basic for short patterns (single course, enforced)
    if len(text.split()) <= 8:
        return "basic"

    return "advanced"


# ------------------------------------------------------
# 2️⃣ Basic parser — handles short and regular forms
# ------------------------------------------------------
def parse_basic(raw_text: str, subject_hint: str, valid_subjects=None):
    """
    Parse simple patterns like:
      - 'Enforced requisite: course 31.'
      - 'Requisites: courses 111, 131.'
      - 'Requisites: Engineering 183EW or 185EW.'
    """

    text = raw_text.strip()
    req_type = "enforced" if text.lower().startswith("enforced") else "unenforced"

    # Remove header labels
    text = re.sub(r"(?i)enforced requisites?:", "", text)
    text = re.sub(r"(?i)requisites?:", "", text)
    text = re.sub(r"(?i)requisite:", "", text)
    text = text.strip(". ").strip()

    # Split into course tokens
    # Examples: ("Computer Science", "32"), ("", "31"), ("Engineering", "183EW")
    pattern = re.compile(
        r'([A-Z][A-Za-z]*(?: [A-Z][A-Za-z]*)*)?\s*(\d+[A-Z]?(?:[A-Z])?)'
    )
    matches = pattern.findall(text)

    if not matches:
        return None

    courses = []
    current_subject = subject_hint

    for subj, num in matches:
        subj = subj.strip() if subj.strip() else current_subject
        inferred = subj.lower() == subject_hint.lower()

        # Validate against subject list (optional)
        if valid_subjects is not None and subj.lower() not in [s.lower() for s in valid_subjects]:
            subj = subject_hint
            inferred = True

        courses.append({
            "subject": subj,
            "number": num,
            "inferred": inferred
        })

    # Group all courses with OR (since basic cases are usually alternatives or lists)
    result = {
        "requisites": [
            {
                "type": req_type,
                "groups": [
                    {
                        "operator": "OR",
                        "courses": courses
                    }
                ]
            }
        ]
    }

    return result


# ------------------------------------------------------
# 3️⃣ Unified entrypoint — used by main.py
# ------------------------------------------------------
def parse_requisites(raw_text: str, subject_hint: str, valid_subjects=None):
    """Top-level entry: classify and route parsing logic."""
    category = classify_requisite(raw_text)

    if category == "basic":
        return parse_basic(raw_text, subject_hint, valid_subjects)

    # For advanced/unhandled patterns → return None so DB remains NULL
    return None
