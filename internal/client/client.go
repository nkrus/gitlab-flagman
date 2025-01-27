package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/nkrus/gitlab-flagman/config"
)

type GitLabClient struct {
	BaseURL    string
	Token      string
	ProjectID  string
	httpClient *http.Client
}

type Pagination struct {
	page       int
	nextPage   int
	prevPage   int
	perPage    int
	totalPages int
	total      int
}

func NewGitLabClient(baseURL, token, projectID string, requestTimeout int) *GitLabClient {
	return &GitLabClient{
		BaseURL:   baseURL,
		Token:     token,
		ProjectID: projectID,
		httpClient: &http.Client{
			Timeout: time.Duration(requestTimeout) * time.Second,
		},
	}
}

const maxConcurrency = 5
const maxPerPage = 2
const xPageHeader = "X-Page"              // The index of the current page (starting at 1).
const xNextPageHeader = "X-Next-Page"     // The index of the next page.
const xPrevPageHeader = "X-Prev-Page"     // The index of the previous page.
const xPerPageHeader = "X-Per-Page"       // The number of items per page.
const xTotalPagesHeader = "X-Total-Pages" // The total number of pages.
const xTotalHeader = "X-Total"            // The total number of items.

func (c *GitLabClient) GetAllFeatureFlags(ctx context.Context) ([]config.FeatureFlag, error) {
	_, pagination, err := c.getFeatureFlagsWithPagination(ctx, 1, maxPerPage)
	if err != nil {
		return nil, err
	}
	// Канал для передачи результатов
	results := make(chan []config.FeatureFlag, pagination.totalPages)
	errors := make(chan error, pagination.totalPages)

	// Семафор для ограничения числа горутин
	semaphore := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	// Запускаем горутины для загрузки всех страниц
	for page := 1; page <= pagination.totalPages; page++ {
		wg.Add(1)
		go func(ctx context.Context, page int) {
			defer wg.Done()
			featureFlags, _, err := c.getFeatureFlagsWithPagination(ctx, page, maxPerPage)
			if err != nil {
				errors <- err
				return
			}
			// Ограничиваем количество одновременных запросов
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			results <- featureFlags
		}(ctx, page)
	}

	// Закрываем канал результатов, когда все горутины завершатся
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// Собираем все результаты
	var allFeatureFlags []config.FeatureFlag
	for result := range results {
		allFeatureFlags = append(allFeatureFlags, result...)
	}

	// Проверяем, были ли ошибки
	if len(errors) > 0 {
		err := <-errors
		return nil, err
	}

	return allFeatureFlags, err
}

func (c *GitLabClient) getFeatureFlagsWithPagination(ctx context.Context, page, perPage int) ([]config.FeatureFlag, Pagination, error) {
	endpoint := fmt.Sprintf("%s/projects/%s/feature_flags?page=%d&per_page=%d", c.BaseURL, c.ProjectID, page, perPage)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, Pagination{}, fmt.Errorf("failed to create GET request: %w", err)
	}
	req.Header.Set("Private-Token", c.Token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, Pagination{}, fmt.Errorf("failed to get feature flags: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, Pagination{}, fmt.Errorf("failed to get feature flags: %s", resp.Status)
	}

	pagination, err := getPagination(resp)
	if err != nil {
		return nil, Pagination{}, err
	}

	var featureFlags []config.FeatureFlag
	if err := json.NewDecoder(resp.Body).Decode(&featureFlags); err != nil {
		return nil, pagination, fmt.Errorf("failed to decode feature flags response: %w", err)
	}
	return featureFlags, pagination, nil
}

func getPagination(resp *http.Response) (Pagination, error) {
	parseHeader := func(header string) (int, error) {
		if header == "" {
			return 0, nil // Если заголовок пустой, возвращаем 0
		}
		return strconv.Atoi(header)
	}

	page, err := parseHeader(resp.Header.Get(xPageHeader))
	if err != nil {
		return Pagination{}, fmt.Errorf("failed to parse %s: %w", xPageHeader, err)
	}

	nextPage, err := parseHeader(resp.Header.Get(xNextPageHeader))
	if err != nil {
		return Pagination{}, fmt.Errorf("failed to parse %s: %w", xNextPageHeader, err)
	}

	prevPage, err := parseHeader(resp.Header.Get(xPrevPageHeader))
	if err != nil {
		return Pagination{}, fmt.Errorf("failed to parse %s: %w", xPrevPageHeader, err)
	}

	perPage, err := parseHeader(resp.Header.Get(xPerPageHeader))
	if err != nil {
		return Pagination{}, fmt.Errorf("failed to parse %s: %w", xPerPageHeader, err)
	}

	totalPages, err := parseHeader(resp.Header.Get(xTotalPagesHeader))
	if err != nil {
		return Pagination{}, fmt.Errorf("failed to parse %s: %w", xTotalPagesHeader, err)
	}

	total, err := parseHeader(resp.Header.Get(xTotalHeader))
	if err != nil {
		return Pagination{}, fmt.Errorf("failed to parse %s: %w", xTotalHeader, err)
	}

	return Pagination{
		page:       page,
		nextPage:   nextPage,
		prevPage:   prevPage,
		perPage:    perPage,
		totalPages: totalPages,
		total:      total,
	}, nil
}

func (c *GitLabClient) DeleteFeatureFlag(ctx context.Context, flagName string) error {
	deleteURL := fmt.Sprintf("%s/projects/%s/feature_flags/%s", c.BaseURL, c.ProjectID, flagName)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL, nil)
	if err != nil {
		return fmt.Errorf("error creating DELETE request: %w", err)
	}
	req.Header.Set("Private-Token", c.Token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error deleting feature flag %s: %w", flagName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error deleting feature flag %s: %s", flagName, resp.Status)
	}

	return nil
}

func (c *GitLabClient) CreateFeatureFlag(ctx context.Context, flag config.FeatureFlag) error {
	createURL := fmt.Sprintf("%s/projects/%s/feature_flags", c.BaseURL, c.ProjectID)

	data, err := json.Marshal(flag)
	if err != nil {
		return fmt.Errorf("failed to marshal feature flag: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, createURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create POST request: %w", err)
	}
	req.Header.Set("Private-Token", c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create feature flag %s: %w", flag.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create feature flag %s: %s", flag.Name, resp.Status)
	}

	return nil
}
