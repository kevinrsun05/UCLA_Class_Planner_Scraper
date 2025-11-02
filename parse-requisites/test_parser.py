from parser import parse_requisites
import json

print(json.dumps(parse_requisites(
    "Requisites: course 32 or Program in Computing 10C; "
    "Civil and Environmental Engineering 110 or Electrical and Computer Engineering 131A or "
    "Mathematics 170A or 170E or Statistics 100A; Mathematics 33A.",
    "COM SCI"
), indent=2))
