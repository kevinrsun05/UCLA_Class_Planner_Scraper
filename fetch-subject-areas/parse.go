package main

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
)

type RawSubjectArea struct {
	Label string `json:"label"` // "Aerospace Studies (AERO ST)"
	Value string `json:"value"` // "AERO ST" (sometimes padded with spaces)
}

type SubjectArea struct {
	Name string // "Aerospace Studies"
	Code string // "AERO ST"
}

func ParseSubjectAreas(body []byte) ([]SubjectArea, error) {
	s := string(body)

	// Find the wrapper: SearchPanelSetup('<JSON-ARRAY-AS-STRING>',
	const prefix = "SearchPanelSetup('"
	start := strings.Index(s, prefix)
	if start == -1 {
		return nil, fmt.Errorf("SearchPanelSetup(' not found")
	}
	start += len(prefix)

	// Find the closing single quote that ends the JSON-string argument
	endRel := strings.IndexByte(s[start:], '\'')
	if endRel == -1 {
		return nil, fmt.Errorf("closing quote after JSON string not found")
	}
	end := start + endRel

	// Extract and unescape the JSON text (HTML entities like &quot;)
	jsonText := html.UnescapeString(s[start:end])
	jsonText = strings.TrimSpace(jsonText)

	// Now this is a real JSON array: [{"label":"...","value":"..."}, ...]
	var raw []RawSubjectArea
	if err := json.Unmarshal([]byte(jsonText), &raw); err != nil {
		return nil, fmt.Errorf("unmarshal subject areas: %w", err)
	}

	out := make([]SubjectArea, 0, len(raw))
	for _, r := range raw {
		name := strings.TrimSpace(r.Label)
		// Strip trailing " (CODE)" from label
		if p := strings.LastIndex(name, " ("); p != -1 && strings.HasSuffix(name, ")") {
			name = name[:p]
		}
		code := strings.TrimSpace(html.UnescapeString(r.Value)) // some codes have padding
		if name == "" || code == "" {
			continue
		}
		out = append(out, SubjectArea{Name: name, Code: code})
	}
	return out, nil
}
