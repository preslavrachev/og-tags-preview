package ogtags

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// modify this might break some test cases
var testBreakerConfig = breakerConfig{
	maxRequest:       3,
	interval:         10 * time.Second,
	timeout:          30 * time.Second,
	tripRequestCount: 5,
	tripFailureRatio: 0.6,
}

func Test_GetOGTags(t *testing.T) {

	t.Run("01_hello.html", func(t *testing.T) {
		url := "https://ogp.me/"
		htmlContent, err := os.ReadFile("testhtml/01_hello.html")
		if err != nil {
			t.Fatalf("Failed to read test file: %v", err)
		}

		mc := &HTTPClientMock{
			GetFunc: func(url string) (*http.Response, error) {
				r := &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(string(htmlContent))),
				}
				return r, nil
			},
		}

		want := &OGTags{
			URL: url,
			Tags: []string{
				"og:title Hello! Open Graph!!",
				"og:type website",
				"og:image /images/01.png",
				"og:image ../images/02.png",
				"og:audio:url https://foobaa.com/hoge.mp3",
				"og:video:url https://foobaa.com/fuga.mp4",
				"og:description This description should be preferred",
				"og:site_name Test Website",
			},
		}

		ogTagsclient := New(mc)
		got, err := ogTagsclient.GetOGTags(url)
		assert.True(t, len(mc.GetCalls()) == 1)
		assert.Nil(t, err)
		assert.Equal(t, got.URL, want.URL)
		assert.Equal(t, got.Tags, want.Tags)
	})

	t.Run("02_haisa.html", func(t *testing.T) {
		url := "https://ogp.me/"
		htmlContent, err := os.ReadFile("testhtml/02_haisa.html")
		if err != nil {
			t.Fatalf("Failed to read test file: %v", err)
		}

		mc := &HTTPClientMock{
			GetFunc: func(url string) (*http.Response, error) {
				r := &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(string(htmlContent))),
				}
				return r, nil
			},
		}

		want := &OGTags{
			URL: url,
			Tags: []string{
				"og:title ",
				"og:description All Genre Music Party",
				"og:type website",
				"og:url https://haisai.party/",
				"og:updated_time 2017-09-07T12:00:00+09:00",
			},
		}

		ogTagsclient := New(mc)
		got, err := ogTagsclient.GetOGTags(url)
		assert.Nil(t, err)
		assert.True(t, len(mc.GetCalls()) == 1)
		assert.Equal(t, got.URL, want.URL)
		assert.Equal(t, got.Tags, want.Tags)
	})

	t.Run("03_image.html", func(t *testing.T) {
		url := "https://ogp.me/"
		htmlContent, err := os.ReadFile("testhtml/03_image.html")
		if err != nil {
			t.Fatalf("Failed to read test file: %v", err)
		}

		mc := &HTTPClientMock{
			GetFunc: func(url string) (*http.Response, error) {
				r := &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(string(htmlContent))),
				}
				return r, nil
			},
		}

		want := &OGTags{
			URL: url,
			Tags: []string{
				"og:image:width parent image is NOT yet defined",
				"og:image //www-cdn.jtvnw.net/images/twitch_logo3.jpg",
			},
		}

		ogTagsclient := New(mc)
		got, err := ogTagsclient.GetOGTags(url)
		assert.Nil(t, err)
		assert.True(t, len(mc.GetCalls()) == 1)
		assert.Equal(t, got.URL, want.URL)
		assert.Equal(t, got.Tags, want.Tags)
	})

	t.Run("04_video.html", func(t *testing.T) {
		url := "https://ogp.me/"
		htmlContent, err := os.ReadFile("testhtml/04_video.html")
		if err != nil {
			t.Fatalf("Failed to read test file: %v", err)
		}

		mc := &HTTPClientMock{
			GetFunc: func(url string) (*http.Response, error) {
				r := &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(string(htmlContent))),
				}
				return r, nil
			},
		}

		want := &OGTags{
			URL:  url,
			Tags: []string{},
		}

		ogTagsclient := New(mc)
		got, err := ogTagsclient.GetOGTags(url)
		assert.Nil(t, err)
		assert.True(t, len(mc.GetCalls()) == 1)
		assert.Equal(t, got.URL, want.URL)
		assert.Equal(t, got.Tags, want.Tags)
	})

	t.Run("05_reddit.html", func(t *testing.T) {
		url := "https://ogp.me/"
		htmlContent, err := os.ReadFile("testhtml/05_reddit.html")
		if err != nil {
			t.Fatalf("Failed to read test file: %v", err)
		}

		mc := &HTTPClientMock{
			GetFunc: func(url string) (*http.Response, error) {
				r := &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(string(htmlContent))),
				}
				return r, nil
			},
		}

		want := &OGTags{
			URL: url,
			Tags: []string{
				"og:image https://www.redditstatic.com/shreddit/assets/favicon/192x192.png",
				"og:image:width 256",
				"og:image:height 256",
				"og:site_name Reddit",
				"og:title reddit",
				"og:ttl 600",
				"og:type website",
				"og:url https://www.reddit.com/",
			},
		}

		ogTagsclient := New(mc)
		got, err := ogTagsclient.GetOGTags(url)
		assert.Nil(t, err)
		assert.True(t, len(mc.GetCalls()) == 1)
		assert.Equal(t, got.URL, want.URL)
		assert.Equal(t, got.Tags, want.Tags)
	})

	t.Run("http client error", func(t *testing.T) {
		url := "https://example.com"

		mc := &HTTPClientMock{
			GetFunc: func(url string) (*http.Response, error) {
				return nil, errors.New("network error")
			},
		}

		ogTagsClient := New(mc)
		got, err := ogTagsClient.GetOGTags(url)
		assert.True(t, len(mc.GetCalls()) == 1)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "GetOGTags:retryablehttp.Get")
	})

	t.Run("circuit breaker opens after consecutive failures", func(t *testing.T) {
		url := "https://example.com"

		mc := &HTTPClientMock{
			GetFunc: func(url string) (*http.Response, error) {
				return nil, errors.New("network error")
			},
		}

		ogTagsClient := New(mc)
		ogTagsClient.bkcfg = testBreakerConfig

		// Make multiple calls to trigger circuit breaker
		for i := 0; i < 5; i++ {
			got, err := ogTagsClient.GetOGTags(url)
			assert.Error(t, err)
			assert.Nil(t, got)
			assert.Contains(t, err.Error(), "GetOGTags:retryablehttp.Get")
		}

		assert.True(t, len(mc.GetCalls()) == 5)

		// Next call should fail immediately due to open circuit breaker
		got, err := ogTagsClient.GetOGTags(url)

		assert.True(t, len(mc.GetCalls()) == 5)
		assert.Error(t, err)
		assert.Nil(t, got)
		// Should contain circuit breaker error message
		assert.Contains(t, err.Error(), "circuit breaker is open")
	})

	t.Run("circuit breaker isolates different hosts", func(t *testing.T) {
		url1 := "https://example.com"
		url2 := "https://different.com"

		mc := &HTTPClientMock{
			GetFunc: func(url string) (*http.Response, error) {
				switch url {
				case url1:
					return nil, errors.New("network error")
				case url2:
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(`<html><head><meta property="og:title" content="Test"/></head></html>`)),
					}, nil
				default:
					return nil, errors.New("unexpected URL")
				}
			},
		}

		ogTagsClient := New(mc)
		ogTagsClient.bkcfg = testBreakerConfig

		// Trigger circuit breaker for first host
		for i := 0; i < 5; i++ {
			got, err := ogTagsClient.GetOGTags(url1)
			assert.Error(t, err)
			assert.Nil(t, got)
		}

		// Verify we made 5 calls to the first host
		url1Calls := 0
		for _, call := range mc.GetCalls() {
			if call.URL == url1 {
				url1Calls++
			}
		}
		assert.Equal(t, 5, url1Calls)

		// Second host should still work
		got, err := ogTagsClient.GetOGTags(url2)
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, url2, got.URL)

		// Verify we made 1 call to the second host
		url2Calls := 0
		for _, call := range mc.GetCalls() {
			if call.URL == url2 {
				url2Calls++
			}
		}
		assert.Equal(t, 1, url2Calls)

		// First host should still be blocked
		got, err = ogTagsClient.GetOGTags(url1)

		assert.Equal(t, 6, len(mc.GetCalls()), "Circuit breaker should prevent additional HTTP calls")
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "circuit breaker is open")
	})

	t.Run("circuit breaker reuses same breaker for same host", func(t *testing.T) {
		url1 := "https://example.com/page1"
		url2 := "https://example.com/page2"

		mc := &HTTPClientMock{
			GetFunc: func(url string) (*http.Response, error) {
				// Both URLs from same host should fail
				if url == url1 || url == url2 {
					return nil, errors.New("network error")
				}
				return nil, errors.New("unexpected URL")
			},
		}

		ogTagsClient := New(mc)
		ogTagsClient.bkcfg = testBreakerConfig

		// Make some failures on first URL
		for i := 0; i < 3; i++ {
			got, err := ogTagsClient.GetOGTags(url1)
			assert.Error(t, err)
			assert.Nil(t, got)
		}

		// Make failures on second URL (same host) - should contribute to same breaker
		for i := 0; i < 2; i++ {
			got, err := ogTagsClient.GetOGTags(url2)
			assert.Error(t, err)
			assert.Nil(t, got)
		}

		// Verify we made the expected number of HTTP calls (5 total)
		assert.Equal(t, 5, len(mc.GetCalls()))

		// Count calls for each URL
		url1Calls := 0
		url2Calls := 0
		for _, call := range mc.GetCalls() {
			switch call.URL {
			case url1:
				url1Calls++
			case url2:
				url2Calls++
			}
		}
		assert.Equal(t, 3, url1Calls)
		assert.Equal(t, 2, url2Calls)

		// Next call to either URL should hit open circuit breaker (no additional HTTP calls)
		got, err := ogTagsClient.GetOGTags(url1)
		assert.Equal(t, 5, len(mc.GetCalls()), "Circuit breaker should prevent HTTP call to url1")
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "circuit breaker")

		got, err = ogTagsClient.GetOGTags(url2)
		assert.Equal(t, 5, len(mc.GetCalls()), "Circuit breaker should prevent HTTP call to url2")
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "circuit breaker")
	})

	t.Run("getHost error", func(t *testing.T) {
		invalidURL := "haha"
		mc := &HTTPClientMock{}
		ogTagsclient := New(mc)
		got, err := ogTagsclient.GetOGTags(invalidURL)
		assert.Equal(t, 0, len(mc.GetCalls()))
		assert.True(t, strings.Contains(err.Error(), "GetOGTags:getHost"))
		assert.Empty(t, got)
	})

	t.Run("circuit breaker cache exists for 1 host", func(t *testing.T) {
		url := "https://ogp.me/"
		host, err := getHost(url)
		if err != nil {
			t.Fatal(err)
		}

		mc := &HTTPClientMock{
			GetFunc: func(url string) (*http.Response, error) {
				r := &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader("")),
				}
				return r, nil
			},
		}

		want := &OGTags{
			URL:  url,
			Tags: []string{},
		}

		ogTagsClient := New(mc)
		ogTagsClient.bkcfg = testBreakerConfig

		assert.Equal(t, 0, ogTagsClient.breakersCache.Len())

		got, err := ogTagsClient.GetOGTags(url)

		assert.True(t, ogTagsClient.breakersCache.Contains(host))

		assert.Nil(t, err)
		assert.Equal(t, got.URL, want.URL)
		assert.Equal(t, got.Tags, want.Tags)
	})

	t.Run("circuit breaker exists for multiple request for same host", func(t *testing.T) {
		url := "https://ogp.me/"
		host, err := getHost(url)
		if err != nil {
			t.Fatal(err)
		}

		mc := &HTTPClientMock{
			GetFunc: func(url string) (*http.Response, error) {
				r := &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader("")),
				}
				return r, nil
			},
		}

		want := &OGTags{
			URL:  url,
			Tags: []string{},
		}

		ogTagsClient := New(mc)
		ogTagsClient.bkcfg = testBreakerConfig

		assert.Equal(t, 0, ogTagsClient.breakersCache.Len())

		var got *OGTags
		for i := 0; i < 10; i++ {
			got, err = ogTagsClient.GetOGTags(url)
		}

		assert.True(t, ogTagsClient.breakersCache.Contains(host))
		assert.Equal(t, 1, ogTagsClient.breakersCache.Len())

		assert.Nil(t, err)
		assert.Equal(t, got.URL, want.URL)
		assert.Equal(t, got.Tags, want.Tags)
	})

	t.Run("circuit breaker exists for multiple host", func(t *testing.T) {
		url1 := "https://ogp.me/"
		url2 := "https://ogp2.me/"
		host, err := getHost(url1)
		if err != nil {
			t.Fatal(err)
		}
		host2, err := getHost(url2)
		if err != nil {
			t.Fatal(err)
		}

		mc := &HTTPClientMock{
			GetFunc: func(url string) (*http.Response, error) {
				switch url {
				case url1:
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader("")),
					}, nil
				case url2:
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader("")),
					}, nil
				default:
					return nil, errors.New("unexpected URL")
				}
			},
		}

		ogTagsClient := New(mc)
		ogTagsClient.bkcfg = testBreakerConfig

		assert.Equal(t, 0, ogTagsClient.breakersCache.Len())

		ogTagsClient.GetOGTags(url1)
		ogTagsClient.GetOGTags(url2)

		assert.Equal(t, 2, ogTagsClient.breakersCache.Len())
		assert.True(t, ogTagsClient.breakersCache.Contains(host))
		assert.True(t, ogTagsClient.breakersCache.Contains(host2))
	})

}
