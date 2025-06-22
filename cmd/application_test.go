package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TrungNNg/og-tag/internal/ogtags"
	"github.com/TrungNNg/og-tag/internal/ogtags_cache"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func Test_ogTagHandler(t *testing.T) {

	t.Run("happy path, new valid url", func(t *testing.T) {
		url := "https://example.com"

		// No cache, successful set
		getCacheCalled := 0
		setCacheCalled := 0
		ogCacheMock := &ogtags_cache.OGCacheClientMock{
			GetFunc: func(url string) (string, error) {
				getCacheCalled++
				return "", ogtags_cache.ErrKeyNotFound
			},
			SetFunc: func(url string, jsonByte []byte) error {
				setCacheCalled++
				return nil
			},
		}

		// sucessful get og tags from url
		getClientCall := 0
		ogsTag := &ogtags.OGTags{
			URL:  url,
			Tags: []string{"og:title example", "og:url https://example.com"},
		}
		ogClientMock := &ogtags.OGTagClientMock{
			GetOGTagsFunc: func(url string) (*ogtags.OGTags, error) {
				getClientCall++
				ogs := ogsTag
				return ogs, nil
			},
		}

		app := &application{
			client:    ogClientMock,
			cache:     ogCacheMock,
			validator: validator.New(),
		}

		ts := httptest.NewServer(app.routes())
		defer ts.Close()

		payload := map[string]string{"url": url}
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := http.Post(ts.URL+"/og", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		got, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		want := httptest.NewRecorder()
		err = app.writeJSON(want, http.StatusOK, envelope{"result": ogsTag}, nil)
		if err != nil {
			t.Fatal(err)
		}

		// check cache calls
		assert.Equal(t, 2, getCacheCalled+setCacheCalled)

		// check client calls
		assert.Equal(t, 1, getClientCall)

		// check resp body
		assert.Equal(t, got, want.Body.Bytes())
	})

	t.Run("cache hit, return cached data", func(t *testing.T) {
		url := "https://cached-example.com"

		cachedResponse := `{
		"result": {
			"url": "https://cached-example.com",
			"tags": ["og:title cached example", "og:url https://cached-example.com"]
		}
	}
`

		getCacheCalled := 0
		setCacheCalled := 0
		ogCacheMock := &ogtags_cache.OGCacheClientMock{
			GetFunc: func(url string) (string, error) {
				getCacheCalled++
				return cachedResponse, nil
			},
			SetFunc: func(url string, jsonByte []byte) error {
				setCacheCalled++
				return nil
			},
		}

		// Client should not be called when cache hits
		getClientCall := 0
		ogClientMock := &ogtags.OGTagClientMock{
			GetOGTagsFunc: func(url string) (*ogtags.OGTags, error) {
				getClientCall++
				return nil, fmt.Errorf("should not be called")
			},
		}

		app := &application{
			client:    ogClientMock,
			cache:     ogCacheMock,
			validator: validator.New(),
		}

		ts := httptest.NewServer(app.routes())
		defer ts.Close()

		payload := map[string]string{"url": url}
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := http.Post(ts.URL+"/og", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		got, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		// Verify cache was called but client was not
		assert.Equal(t, 1, getCacheCalled)
		assert.Equal(t, 0, setCacheCalled)
		assert.Equal(t, 0, getClientCall)

		// Verify response matches cached data
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
		assert.JSONEq(t, cachedResponse, string(got))
	})

	t.Run("invalid URL validation error", func(t *testing.T) {
		invalidURL := "not-a-valid-url"

		app := &application{
			validator: validator.New(),
		}

		ts := httptest.NewServer(app.routes())
		defer ts.Close()

		payload := map[string]string{"url": invalidURL}
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := http.Post(ts.URL+"/og", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("cache miss, client fetch fails", func(t *testing.T) {
		url := "https://failing-example.com"

		getCacheCalled := 0
		setCacheCalled := 0
		ogCacheMock := &ogtags_cache.OGCacheClientMock{
			GetFunc: func(url string) (string, error) {
				getCacheCalled++
				return "", ogtags_cache.ErrKeyNotFound
			},
			SetFunc: func(url string, jsonByte []byte) error {
				setCacheCalled++
				return nil
			},
		}

		getClientCall := 0
		ogClientMock := &ogtags.OGTagClientMock{
			GetOGTagsFunc: func(url string) (*ogtags.OGTags, error) {
				getClientCall++
				return nil, fmt.Errorf("failed to fetch OG tags")
			},
		}

		app := &application{
			client:    ogClientMock,
			cache:     ogCacheMock,
			validator: validator.New(),
		}

		ts := httptest.NewServer(app.routes())
		defer ts.Close()

		payload := map[string]string{"url": url}
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := http.Post(ts.URL+"/og", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		// Verify calls were made but cache was not set due to client failure
		assert.Equal(t, 1, getCacheCalled)
		assert.Equal(t, 0, setCacheCalled) // Should not cache on error
		assert.Equal(t, 1, getClientCall)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("cache error (not key not found), still proceed to fetch", func(t *testing.T) {
		url := "https://cache-error-example.com"

		getCacheCalled := 0
		setCacheCalled := 0
		ogCacheMock := &ogtags_cache.OGCacheClientMock{
			GetFunc: func(url string) (string, error) {
				getCacheCalled++
				return "", fmt.Errorf("cache connection error")
			},
			SetFunc: func(url string, jsonByte []byte) error {
				setCacheCalled++
				return nil
			},
		}

		getClientCall := 0
		ogsTag := &ogtags.OGTags{
			URL:  url,
			Tags: []string{"og:title cache error example", "og:url https://cache-error-example.com"},
		}
		ogClientMock := &ogtags.OGTagClientMock{
			GetOGTagsFunc: func(url string) (*ogtags.OGTags, error) {
				getClientCall++
				return ogsTag, nil
			},
		}

		app := &application{
			client:    ogClientMock,
			cache:     ogCacheMock,
			validator: validator.New(),
		}

		ts := httptest.NewServer(app.routes())
		defer ts.Close()

		payload := map[string]string{"url": url}
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := http.Post(ts.URL+"/og", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		got, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		// Verify cache was called, client was called, and response is successful
		assert.Equal(t, 1, getCacheCalled)
		assert.Equal(t, 1, setCacheCalled) // Should still try to cache the result
		assert.Equal(t, 1, getClientCall)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		want := httptest.NewRecorder()
		err = app.writeJSON(want, http.StatusOK, envelope{"result": ogsTag}, nil)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, got, want.Body.Bytes())
	})

	t.Run("successful fetch but cache set fails", func(t *testing.T) {
		url := "https://cache-set-fail-example.com"

		getCacheCalled := 0
		setCacheCalled := 0
		ogCacheMock := &ogtags_cache.OGCacheClientMock{
			GetFunc: func(url string) (string, error) {
				getCacheCalled++
				return "", ogtags_cache.ErrKeyNotFound
			},
			SetFunc: func(url string, jsonByte []byte) error {
				setCacheCalled++
				return fmt.Errorf("cache set failed")
			},
		}

		getClientCall := 0
		ogsTag := &ogtags.OGTags{
			URL:  url,
			Tags: []string{"og:title cache set fail", "og:url https://cache-set-fail-example.com"},
		}
		ogClientMock := &ogtags.OGTagClientMock{
			GetOGTagsFunc: func(url string) (*ogtags.OGTags, error) {
				getClientCall++
				return ogsTag, nil
			},
		}

		app := &application{
			client:    ogClientMock,
			cache:     ogCacheMock,
			validator: validator.New(),
		}

		ts := httptest.NewServer(app.routes())
		defer ts.Close()

		payload := map[string]string{"url": url}
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := http.Post(ts.URL+"/og", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		got, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		// Should still return successful response even if caching fails
		assert.Equal(t, 1, getCacheCalled)
		assert.Equal(t, 1, setCacheCalled)
		assert.Equal(t, 1, getClientCall)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		want := httptest.NewRecorder()
		err = app.writeJSON(want, http.StatusOK, envelope{"result": ogsTag}, nil)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, got, want.Body.Bytes())
	})

	t.Run("empty OG tags response", func(t *testing.T) {
		url := "https://empty-tags-example.com"

		ogCacheMock := &ogtags_cache.OGCacheClientMock{
			GetFunc: func(url string) (string, error) {
				return "", ogtags_cache.ErrKeyNotFound
			},
			SetFunc: func(url string, jsonByte []byte) error {
				return nil
			},
		}

		ogsTag := &ogtags.OGTags{
			URL:  url,
			Tags: []string{}, // Empty tags
		}
		ogClientMock := &ogtags.OGTagClientMock{
			GetOGTagsFunc: func(url string) (*ogtags.OGTags, error) {
				return ogsTag, nil
			},
		}

		app := &application{
			client:    ogClientMock,
			cache:     ogCacheMock,
			validator: validator.New(),
		}

		ts := httptest.NewServer(app.routes())
		defer ts.Close()

		payload := map[string]string{"url": url}
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := http.Post(ts.URL+"/og", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		got, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		want := httptest.NewRecorder()
		err = app.writeJSON(want, http.StatusOK, envelope{"result": ogsTag}, nil)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, got, want.Body.Bytes())
	})

	t.Run("same URL twice - first caches, second returns from cache", func(t *testing.T) {
		url := "https://example.com"

		// Track cache state
		var cachedData []byte
		getCacheCalled := 0
		setCacheCalled := 0
		ogCacheMock := &ogtags_cache.OGCacheClientMock{
			GetFunc: func(url string) (string, error) {
				getCacheCalled++
				if cachedData == nil {
					return "", ogtags_cache.ErrKeyNotFound
				}
				return string(cachedData), nil
			},
			SetFunc: func(url string, jsonByte []byte) error {
				setCacheCalled++
				cachedData = jsonByte
				return nil
			},
		}

		// Client should only be called once
		getClientCall := 0
		ogsTag := &ogtags.OGTags{
			URL:  url,
			Tags: []string{"og:title example", "og:url https://example.com"},
		}
		ogClientMock := &ogtags.OGTagClientMock{
			GetOGTagsFunc: func(url string) (*ogtags.OGTags, error) {
				getClientCall++
				return ogsTag, nil
			},
		}

		app := &application{
			client:    ogClientMock,
			cache:     ogCacheMock,
			validator: validator.New(),
		}

		ts := httptest.NewServer(app.routes())
		defer ts.Close()

		payload := map[string]string{"url": url}
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}

		// First request - should cache miss, fetch from client, then cache result
		resp1, err := http.Post(ts.URL+"/og", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		defer resp1.Body.Close()

		got1, err := io.ReadAll(resp1.Body)
		if err != nil {
			t.Fatal(err)
		}

		// Verify first request behavior
		assert.Equal(t, http.StatusOK, resp1.StatusCode)
		assert.Equal(t, 1, getCacheCalled) // Cache miss
		assert.Equal(t, 1, setCacheCalled) // Cache set
		assert.Equal(t, 1, getClientCall)  // Client called

		// Second request with same URL - should return from cache
		body2, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}

		resp2, err := http.Post(ts.URL+"/og", "application/json", bytes.NewReader(body2))
		if err != nil {
			t.Fatal(err)
		}
		defer resp2.Body.Close()

		got2, err := io.ReadAll(resp2.Body)
		if err != nil {
			t.Fatal(err)
		}

		// Verify second request behavior
		assert.Equal(t, http.StatusOK, resp2.StatusCode)
		assert.Equal(t, 2, getCacheCalled) // Cache hit
		assert.Equal(t, 1, setCacheCalled) // No additional cache set
		assert.Equal(t, 1, getClientCall)  // Client not called again

		// Most importantly: both responses should be identical
		assert.Equal(t, got1, got2, "First and second responses should be identical")

		// Verify both responses match expected format
		want := httptest.NewRecorder()
		err = app.writeJSON(want, http.StatusOK, envelope{"result": ogsTag}, nil)
		if err != nil {
			t.Fatal(err)
		}

		// Both responses should match the expected JSON structure
		assert.Equal(t, want.Body.Bytes(), got1)
		assert.Equal(t, want.Body.Bytes(), got2)
	})

}
