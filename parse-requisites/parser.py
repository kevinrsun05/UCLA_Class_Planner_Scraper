import re
import json

def parse_requisites(raw_text, subject_hint):
    results = []
    req_type = "unenforced"
    text = raw_text.strip()

    if text.lower().startswith("enforced"):
        req_type = "enforced"
        text = re.sub(r"(?i)enforced requisites?:", "", text)
    else:
        text = re.sub(r"(?i)requisites?:", "", text)

    text = re.sub(r'with grade of [A-F][+-]? or better', '', text, flags=re.I)
    text = re.sub(r'courses?', 'course', text, flags=re.I)
    text = re.sub(r'\s+', ' ', text).strip()

    groups = []
    for and_group in re.split(r'\band\b', text, flags=re.I):
        or_courses = []
        for or_part in re.split(r'\bor\b', and_group, flags=re.I):
            matches = re.findall(r'([A-Z][A-Z& ]+)?\s*(\d+[A-Z]?)', or_part)
            for subj, num in matches:
                subj = subj.strip() if subj.strip() else subject_hint
                or_courses.append({
                    "subject": subj,
                    "number": num,
                    "inferred": subj == subject_hint
                })
        if or_courses:
            groups.append({"operator": "OR", "courses": or_courses})

    if groups:
        results.append({"type": req_type, "groups": groups})
    return {"requisites": results}
