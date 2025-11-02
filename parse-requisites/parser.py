import re
import json

# ------------------------------------------------------
# Classifier — decide which parser to use
# ------------------------------------------------------
def classify_requisite(raw_text: str) -> str:
    """Classify a requisite string as 'basic' or 'advanced'."""
    text = raw_text.lower()

    if "one course from" in text or "and one course from" in text:
        return "basic2"  # handle complex but structured 'one course from' cases
    
    # Keywords that indicate complex logic (handled later by advanced parser)
    if any(
        kw in text
        for kw in ["either", "both", "must also", "may be taken concurrently"]
    ):
        return "advanced"

    # Simple course lists or single-course enforcement
    if re.search(r"\bcourses?\b", text) or re.search(r"\bor\b", text):
        return "basic"

    # Default to basic for short patterns (single course, enforced)
    if len(text.split()) <= 8:
        return "basic"

    return "advanced"

def _normalize_subject(subj, subject_hint, valid_subjects):
    subj = subj.strip()
    if not subj:
        return subject_hint, True
    subj_lower = subj.lower()

    if valid_subjects:
        for valid in valid_subjects:
            if subj_lower == valid.lower() or subj_lower in valid.lower():
                return valid, False
    # fallback
    return subject_hint, True

# ------------------------------------------------------
#  Basic parser — handles short and regular forms
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
        subj, inferred = _normalize_subject(subj, subject_hint, valid_subjects)

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
# Basic2 parser — handles “and one course from …” etc.
# ------------------------------------------------------
def parse_basic2(raw_text: str, subject_hint: str, valid_subjects=None):
    text = raw_text.strip()
    req_type = "enforced" if text.lower().startswith("enforced") else "unenforced"

    # remove heading and grading references
    text = re.sub(r"(?i)enforced requisites?:", "", text)
    text = re.sub(r"(?i)requisites?:", "", text)
    text = re.sub(r"(?i)requisite:", "", text)
    text = re.sub(r"with grade of [A-F][+-]? or better", "", text, flags=re.I)
    text = re.sub(r"\s+", " ", text).strip(". ")

    # split into coarse AND groups
    # commas or semicolons separate subclauses, but preserve "or" within each
    parts = re.split(r"\band one course from\b|;", text, flags=re.I)
    groups = []

    course_pattern = re.compile(
        r'([A-Z][A-Za-z]*(?: (?:and|in|of|for|to|the|and the|and of|and in|and for|and to|and&) [A-Z][A-Za-z]*)*)?\s*(\d+[A-Z]?)'
    )

    for part in parts:
        if not part.strip():
            continue
        subtext = part.strip()

        # split on "or" to form OR sets
        or_segments = re.split(r"\bor\b", subtext, flags=re.I)
        current_subject = subject_hint
        courses = []

        for seg in or_segments:
            seg = seg.strip()
            matches = course_pattern.findall(seg)

            for subj, num in matches:
                subj, inferred = _normalize_subject(subj, subject_hint, valid_subjects)
                current_subject = subj
                courses.append({
                    "subject": subj,
                    "number": num,
                    "inferred": inferred
                })

        if courses:
            groups.append({"operator": "OR", "courses": courses})

    if not groups:
        return None

    return {"requisites": [{"type": req_type, "groups": groups}]}

# ------------------------------------------------------
# Unified entrypoint — used by main.py
# ------------------------------------------------------
def parse_requisites(raw_text: str, subject_hint: str, valid_subjects=None):
    """Top-level entry: classify and route parsing logic."""
    category = classify_requisite(raw_text)

    if category == "basic":
        return parse_basic(raw_text, subject_hint, valid_subjects)
    if category == "basic2":
        return parse_basic2(raw_text, subject_hint, valid_subjects)
    # For advanced/unhandled patterns → return None so DB remains NULL
    return None
