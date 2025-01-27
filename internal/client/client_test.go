package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nkrus/gitlab-flagman/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetExistingFeatureFlags(t *testing.T) {
	flags := []config.FeatureFlag{
		{Name: "flag1", Description: "Test flag 1", Active: true},
		{Name: "flag2", Description: "Test flag 2", Active: false},
		{Name: "flag3", Description: "Test flag 3", Active: true},
		{Name: "flag4", Description: "Test flag 4", Active: false},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/projects/1/feature_flags", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Private-Token"), "Private-Token header must be present")
		switch r.URL.RawQuery {
		case "page=1&per_page=2":
			respBytes, err := json.Marshal(flags[:2])
			if err != nil {
				t.Fatalf("Failed to marshal test flags: %v", err)
			}
			w.Header().Set(xPageHeader, "1")
			w.Header().Set(xNextPageHeader, "")
			w.Header().Set(xPrevPageHeader, "1")
			w.Header().Set(xPerPageHeader, "2")
			w.Header().Set(xTotalPagesHeader, "2")
			w.Header().Set(xTotalHeader, "4")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(respBytes)
			if err != nil {
				t.Fatalf("Failed to write response: %v", err)
			}
		case "page=2&per_page=2":
			respBytes, err := json.Marshal(flags[2:])
			if err != nil {
				t.Fatalf("Failed to marshal test flags: %v", err)
			}
			w.Header().Set(xPageHeader, "2")
			w.Header().Set(xNextPageHeader, "2")
			w.Header().Set(xPrevPageHeader, "")
			w.Header().Set(xPerPageHeader, "2")
			w.Header().Set(xTotalPagesHeader, "2")
			w.Header().Set(xTotalHeader, "4")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(respBytes)
			if err != nil {
				t.Fatalf("Failed to write response: %v", err)
			}
		}
	}))
	defer server.Close()

	client := &GitLabClient{
		BaseURL:   server.URL,
		Token:     "some-token",
		ProjectID: "1",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		flags, err := client.GetAllFeatureFlags(ctx)
		assert.NoError(t, err)
		assert.Len(t, flags, 4)
		flagNames := []string{"flag1", "flag2", "flag3", "flag4"}
		for _, v := range flags {
			assert.Contains(t, flagNames, v.Name)
		}
	})

	t.Run("error_creating_request", func(t *testing.T) {
		client.BaseURL = "://invalid-url"

		ctx := context.Background()
		flags, err := client.GetAllFeatureFlags(ctx)
		assert.Error(t, err)
		assert.Nil(t, flags)
	})

	t.Run("error_sending_request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client.BaseURL = server.URL
		ctx := context.Background()
		flags, err := client.GetAllFeatureFlags(ctx)
		assert.Error(t, err)
		assert.Nil(t, flags)
	})

	t.Run("non_ok_status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client.BaseURL = server.URL
		ctx := context.Background()
		flags, err := client.GetAllFeatureFlags(ctx)
		assert.Error(t, err)
		assert.Nil(t, flags)
		assert.Contains(t, err.Error(), "failed to get feature flags: 404")
	})

	t.Run("error_decoding_response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("{ invalid-json }"))
			require.NoError(t, err)
		}))
		defer server.Close()

		client.BaseURL = server.URL
		ctx := context.Background()
		flags, err := client.GetAllFeatureFlags(ctx)
		assert.Error(t, err)
		assert.Nil(t, flags)
		assert.Contains(t, err.Error(), "failed to decode feature flags response")
	})

	t.Run("request_canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		flags, err := client.GetAllFeatureFlags(ctx)
		assert.Error(t, err)
		assert.Nil(t, flags)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestDeleteFeatureFlag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/projects/1/feature_flags/flag1", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Private-Token"), "Private-Token header must be present")

		if r.URL.Path == "/projects/1/feature_flags/flag1" {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := &GitLabClient{
		BaseURL:   server.URL,
		Token:     "some-token",
		ProjectID: "1",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		err := client.DeleteFeatureFlag(ctx, "flag1")
		assert.NoError(t, err)
	})

	t.Run("error_creating_request", func(t *testing.T) {
		client.BaseURL = "://invalid-url"

		ctx := context.Background()
		err := client.DeleteFeatureFlag(ctx, "flag1")
		assert.Error(t, err)
	})

	t.Run("error_sending_request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client.BaseURL = server.URL
		ctx := context.Background()
		err := client.DeleteFeatureFlag(ctx, "flag1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error deleting feature flag flag1")
	})

	t.Run("non_ok_status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client.BaseURL = server.URL
		ctx := context.Background()
		err := client.DeleteFeatureFlag(ctx, "flag1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error deleting feature flag flag1: 400")
	})

	t.Run("flag_not_found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound) // Принудительная ошибка 404
		}))
		defer server.Close()

		client.BaseURL = server.URL
		ctx := context.Background()
		err := client.DeleteFeatureFlag(ctx, "flag1")
		assert.NoError(t, err)
	})

	t.Run("request_canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := client.DeleteFeatureFlag(ctx, "flag1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestCreateFeatureFlag(t *testing.T) {
	mockFlag := config.FeatureFlag{
		Name:        "test-flag",
		Description: "A test feature flag",
		Active:      true,
	}

	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/projects/1/feature_flags", r.URL.Path)
			assert.NotEmpty(t, r.Header.Get("Private-Token"), "Private-Token header must be present")
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var receivedFlag config.FeatureFlag
			err := json.NewDecoder(r.Body).Decode(&receivedFlag)
			require.NoError(t, err)
			assert.Equal(t, mockFlag, receivedFlag)

			w.WriteHeader(http.StatusCreated)
		}))
		defer server.Close()

		client := &GitLabClient{
			BaseURL:   server.URL,
			Token:     "some-token",
			ProjectID: "1",
			httpClient: &http.Client{
				Timeout: 10 * time.Second,
			},
		}

		err := client.CreateFeatureFlag(context.Background(), mockFlag)
		assert.NoError(t, err)
	})

	t.Run("request_creation_error", func(t *testing.T) {
		client := &GitLabClient{
			BaseURL:   "://invalid-url",
			Token:     "some-token",
			ProjectID: "1",
		}

		err := client.CreateFeatureFlag(context.Background(), mockFlag)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create POST request")
	})

	t.Run("request_execution_error", func(t *testing.T) {
		client := &GitLabClient{
			BaseURL:   "http://localhost:12345",
			Token:     "some-token",
			ProjectID: "1",
			httpClient: &http.Client{
				Timeout: 1 * time.Second,
			},
		}

		err := client.CreateFeatureFlag(context.Background(), mockFlag)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create feature flag test-flag")
	})

	t.Run("server_error_status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client := &GitLabClient{
			BaseURL:   server.URL,
			Token:     "some-token",
			ProjectID: "1",
			httpClient: &http.Client{
				Timeout: 10 * time.Second,
			},
		}

		err := client.CreateFeatureFlag(context.Background(), mockFlag)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create feature flag test-flag: 400 Bad Request")
	})
}
