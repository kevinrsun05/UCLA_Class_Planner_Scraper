package main

import (
	"encoding/json"
	"regexp"
	"strings"
)

type CourseDetail struct {
	Title       string `json:"course_title"`
	Description string `json:"crs_desc"`
	Units       string `json:"unt_rng"`
}

type Course struct {
	Title          string
	Number         string
	Description    string
	Units          string
	RequisitesText string // Raw requisite text extracted from description
}

// ParseCourseDescriptions parses the JSON array returned by the API into []Course
func ParseCourseDescriptions(subjectArea SubjectArea, data []byte) ([]Course, error) {
	var raw []CourseDetail
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}

	var courses []Course
	for _, d := range raw {
		number, title := splitTitle(d.Title)
		requisitesText := ExtractRequisitesText(d.Description)

		courses = append(courses, Course{
			Title:          title,
			Number:         number,
			Description:    d.Description,
			Units:          d.Units,
			RequisitesText: requisitesText,
		})
	}
	return courses, nil
}

// ExtractRequisitesText extracts the full requisite text from description
func ExtractRequisitesText(desc string) string {
	desc = strings.ReplaceAll(desc, "\r\n", " ")
	// Match sentences with requisite/prerequisite/corequisite
	re := regexp.MustCompile(`(?i)(Enforced )?(Co)?requisite[s]?:[^.]+\.`)
	matches := re.FindAllString(desc, -1)
	return strings.Join(matches, " ")
}

func splitTitle(raw string) (number, title string) {
	parts := strings.SplitN(raw, ". ", 2)
	if len(parts) == 2 {
		number = strings.TrimSpace(parts[0])
		title = strings.TrimSpace(parts[1])
	} else {
		number = raw
		title = ""
	}
	return
}
