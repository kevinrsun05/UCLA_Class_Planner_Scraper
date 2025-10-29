package main

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SubjectArea struct {
	ID   int64
	Code string
	Name string
}

type Course struct {
	ID            int64
	SubjectAreaID int64
	Title         string
	Number        string // e.g., "111"
}

var (
	firstPageModelRegex = regexp.MustCompile(`(SearchPanel\.SearchData = JSON\.stringify\()({.*})`)
	headerRegex         = regexp.MustCompile(`(\S*) - (.*)`) // number - title (fallback path)
)

var (
	errNoResults   = errors.New("no results found")
	errNoPageCount = errors.New("no pageCount on first page")
)

// ParseFirstPage parses the first page, fills courseMap, and returns pageCount + model token.
func ParseFirstPage(subjectAreaID int64, doc *goquery.Document, courseMap map[string]*Course) (pageCount int, modelToken string, err error) {
	// No results?
	if doc.Find("#spanNoSearchResults").Length() != 0 {
		return 0, "", errNoResults
	}

	ParseCourses(subjectAreaID, doc, courseMap)

	pageCountStr, exists := doc.Find("#pageCount").Attr("value")
	if !exists {
		return 0, "", errNoPageCount
	}
	pageCount, err = strconv.Atoi(pageCountStr)
	if err != nil {
		return 0, "", errNoPageCount
	}

	// Extract the serialized model token for pagination
	if v, ok := doc.Find("input[name='model']").Attr("value"); ok && strings.TrimSpace(v) != "" {
		modelToken = v
	} else {
		body, _ := goquery.OuterHtml(doc.Selection)
		matches := firstPageModelRegex.FindStringSubmatch(body)
		if len(matches) == 3 {
			modelToken = matches[2]
		}
	}
	return pageCount, modelToken, nil
}

// ParseCourses parses all the courses on a given page and adds them to the course map.
func ParseCourses(subjectAreaID int64, doc *goquery.Document, courseMap map[string]*Course) {
	// Preferred DOM structure
	doc.Find(".row-fluid.class-title").Each(func(i int, s *goquery.Selection) {
		num := strings.TrimSpace(s.Find(".primary-row .column-one").Text())
		title := strings.TrimSpace(s.Find(".primary-row .column-two").Text())
		if num == "" || title == "" {
			return
		}
		key := num + "-" + title
		if _, ok := courseMap[key]; ok {
			return
		}
		courseMap[key] = &Course{
			SubjectAreaID: subjectAreaID,
			Title:         title,
			Number:        num,
		}
	})

	// Fallback DOM structure if needed
	if len(courseMap) == 0 {
		results := doc.Find("#resultsTitle")
		links := results.Find("h3 > button")
		for i := range links.Nodes {
			link := strings.TrimSpace(links.Eq(i).Text())
			header := headerRegex.FindStringSubmatch(link) // [full, number, title]
			if len(header) < 3 {
				continue
			}
			num := strings.TrimSpace(header[1])
			title := strings.TrimSpace(header[2])
			if num == "" || title == "" {
				continue
			}
			key := num + "-" + title
			if _, ok := courseMap[key]; ok {
				continue
			}
			courseMap[key] = &Course{
				SubjectAreaID: subjectAreaID,
				Title:         title,
				Number:        num,
			}
		}
	}
}

func FetchAndParseCourses(ctx context.Context, subjectArea SubjectArea, term string) ([]Course, error) {
	firstDoc, err := FetchFirstPage(ctx, subjectArea.Code, term)
	if err != nil {
		return nil, err
	}

	courseMap := make(map[string]*Course)
	pageCount, modelToken, err := ParseFirstPage(subjectArea.ID, firstDoc, courseMap)
	if err != nil {
		return nil, err
	}

	if pageCount > 1 && modelToken != "" {
		for page := 2; page <= pageCount; page++ {
			doc, err := FetchAdditionalPage(ctx, modelToken, page)
			if err != nil {
				continue // tolerate a failed page
			}
			ParseCourses(subjectArea.ID, doc, courseMap)
		}
	}

	out := make([]Course, 0, len(courseMap))
	for _, c := range courseMap {
		out = append(out, *c)
	}
	return out, nil
}
