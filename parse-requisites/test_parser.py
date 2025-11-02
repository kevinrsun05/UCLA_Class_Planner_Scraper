# test_parser.py
import json
from parser import parse_requisites

def test_parser():
    valid_subjects = [
        "Computer Science",
        "Program in Computing",
        "Civil and Environmental Engineering",
        "Electrical and Computer Engineering",
        "Mathematics",
        "Statistics",
    ]

    sample = """Requisites: course 32 or Program in Computing 10C with grade of C- or better,
    Mathematics 33A, and one course from Civil Engineering 110,
    Electrical and Computer Engineering 131A, Mathematics 170A,
    Mathematics 170E, or Statistics 100A."""

    parsed = parse_requisites(sample, "Computer Science", valid_subjects)
    print(json.dumps(parsed, indent=2))

if __name__ == "__main__":
    test_parser()
