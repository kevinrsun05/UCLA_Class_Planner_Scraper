import re
import json

def parse_requisites(raw_text: str, subject_hint: str, valid_subjects=None):
    """
    Parse prerequisite strings like:
    "Requisites: course 32 or Program in Computing 10C; Civil and Environmental Engineering 110 or
    Electrical and Computer Engineering 131A or Mathematics 170A or 170E or Statistics 100A; Mathematics 33A."
    """

    results = []
    req_type = "unenforced"
    text = raw_text.strip()

    # --- Detect enforced/unenforced ---
    if text.lower().startswith("enforced"):
        req_type = "enforced"
        text = re.sub(r"(?i)enforced requisites?:", "", text)
    else:
        text = re.sub(r"(?i)\b(requisites?):", "", text)

    # --- Clean ---
    text = re.sub(r"with grade of [A-F][+-]? or better", "", text, flags=re.I)
    text = re.sub(r"\bcourses?\b", "course", text, flags=re.I)
    text = re.sub(r"\s+", " ", text).strip()

    # --- Split into semicolon-delimited requirement blocks (AND groups) ---
    blocks = [b.strip() for b in text.split(";") if b.strip()]
    groups = []

    course_pattern = re.compile(
        r'([A-Z][A-Za-z]*(?: (?:and|in|of|for|to|the|&|and the|and of|and in|and for|and to|and&) [A-Z][A-Za-z]*)*)?\s*(\d+[A-Z]?)'
    )

    for block in blocks:
        or_parts = re.split(r"\bor\b", block, flags=re.I)
        current_subject = None
        courses = []

        for part in or_parts:
            part = part.strip()
            matches = course_pattern.findall(part)

            for subj, num in matches:
                subj = subj.strip() if subj.strip() else current_subject or subject_hint
                inferred = subj == subject_hint
                current_subject = subj

                # Normalize capitalization
                subj = re.sub(r"\s+", " ", subj).strip()
                subj = subj.title() if subj.isupper() else subj

                # ✅ Optional DB validation
                if valid_subjects is not None:
                    if subj.lower() not in valid_subjects:
                        # fallback to hint if invalid
                        subj = subject_hint
                        inferred = True

                courses.append({
                    "subject": subj,
                    "number": num,
                    "inferred": inferred
                })

        if courses:
            groups.append({"operator": "OR", "courses": courses})

    if groups:
        results.append({"type": req_type, "groups": groups})

    return {"requisites": results}
