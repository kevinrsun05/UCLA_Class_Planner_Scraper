package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const baseURL = "https://api.ucla.edu/sis/publicapis/course/getcoursedetail"

type SubjectArea struct {
	ID   int64
	Code string
	Name string
}

func FetchCourseDescriptions(ctx context.Context, subjectArea SubjectArea) ([]byte, error) {
	endpoint, _ := url.Parse(baseURL)
	query := endpoint.Query()
	query.Set("subjectarea", subjectArea.Code)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
