package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	firstPageURL      = "https://sa.ucla.edu/ro/Public/SOC/Results"
	additionalPageURL = "https://sa.ucla.edu/ro/Public/SOC/Results/CourseTitlesView"
)

// FetchFirstPage gets the initial results page for a subject area.
func FetchFirstPage(ctx context.Context, subjectAreaCode, term string) (*goquery.Document, error) {
	u, _ := url.Parse(firstPageURL)
	q := u.Query()
	q.Set("t", term)
	q.Set("sBy", "subject")
	q.Set("subj", subjectAreaCode)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse first page: %w", err)
	}
	return doc, nil
}

// FetchAdditionalPage gets a subsequent page using the model token.
func FetchAdditionalPage(ctx context.Context, model string, pageNumber int) (*goquery.Document, error) {
	u, _ := url.Parse(additionalPageURL)
	q := u.Query()
	q.Set("model", model)
	q.Set("search_by", "subject")
	q.Set("filterFlags", "0000000") // default flags used by site
	q.Set("pageNumber", strconv.Itoa(pageNumber))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	// Required to avoid 404
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch page %d: %w", pageNumber, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status (page %d): %s", pageNumber, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse additional page %d: %w", pageNumber, err)
	}
	return doc, nil
}
