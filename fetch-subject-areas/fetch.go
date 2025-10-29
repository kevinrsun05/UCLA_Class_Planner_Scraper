package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const subjectAreasURL = "https://sa.ucla.edu/ro/ClassSearch/Public/Search/GetSimpleSearchData"

func FetchSubjectAreas(ctx context.Context, term string) ([]byte, error) {
	u, _ := url.Parse(subjectAreasURL)
	q := u.Query()
	q.Set("term_cd", term) // e.g., "25F"
	q.Set("search_type", "subject")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}
