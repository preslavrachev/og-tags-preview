package ogtags

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/sony/gobreaker/v2"
	"golang.org/x/net/html"
)

type OGTags struct {
	URL  string   `json:"url"`
	Tags []string `json:"og_tags"`
}

type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

type OGTagClient interface {
	GetOGTags(url string) (*OGTags, error)
}

type Client struct {
	client        HTTPClient
	breakersCache *lru.Cache[string, *gobreaker.CircuitBreaker[*OGTags]]
	bkcfg         breakerConfig
}

type breakerConfig struct {
	maxRequest       int
	interval         time.Duration
	timeout          time.Duration
	tripRequestCount int
	tripFailureRatio float64
}

func New(c HTTPClient) *Client {
	// One circuit breaker per hostname.
	cache, err := lru.New[string, *gobreaker.CircuitBreaker[*OGTags]](100)
	if err != nil {
		slog.Error("could not create lru cache for circuit breakers")
		os.Exit(1)
	}

	// Configure circuit breaker with the following logic:
	// - Allow up to maxRequest (3) concurrent requests in half-open state.
	// - If at least tripRequestCount (5) requests have occurred,
	//   and tripFailureRatio (60%) or more have failed,
	//   then the breaker will trip (open).
	// - Interval defines how long to wait with no requests before resetting failure stats (10s).
	// - Timeout is how long the breaker stays open before trying again (30s).
	bkcfg := breakerConfig{
		maxRequest:       3,
		interval:         10 * time.Second,
		timeout:          30 * time.Second,
		tripRequestCount: 5,
		tripFailureRatio: 0.6,
	}

	return &Client{
		client:        c,
		breakersCache: cache,
		bkcfg:         bkcfg,
	}
}

func (c *Client) GetOGTags(url string) (*OGTags, error) {
	// get host name from url
	host, err := getHost(url)
	if err != nil || host == "" {
		return nil, fmt.Errorf("GetOGTags:getHost %w", err)
	}

	// check if there is a circuit breaker for this host name is lru cache
	cb, ok := c.breakersCache.Get(host)
	if !ok {
		cb = newHostBreaker(host, c.bkcfg)
		c.breakersCache.Add(host, cb)
	}

	return cb.Execute(func() (*OGTags, error) {
		res, err := c.client.Get(url)
		if err != nil {
			return nil, fmt.Errorf("GetOGTags:retryablehttp.Get %w", err)
		}
		defer res.Body.Close()

		doc, err := html.Parse(res.Body)
		if err != nil {
			return nil, fmt.Errorf("GetOGTags:html.Parse %w", err)
		}

		ogs := &OGTags{
			URL:  url,
			Tags: []string{},
		}

		// only get og tags in meta tags in first head tag
		for n := range doc.Descendants() {
			if n.Type == html.ElementNode && n.Data == "head" {
				for mn := range n.Descendants() {
					if mn.Type == html.ElementNode && mn.Data == "meta" {
						if p, ok := processMetaTag(mn); ok {
							ogs.Tags = append(ogs.Tags, p)
						}
					}
				}
				break
			}
		}
		return ogs, nil
	})
}

func processMetaTag(n *html.Node) (string, bool) {
	var prop, cont string
	for _, attr := range n.Attr {
		switch attr.Key {
		case "property":
			prop = attr.Val
		case "content":
			cont = attr.Val
		}
	}

	if !strings.HasPrefix(prop, "og:") {
		return "", false
	}

	return fmt.Sprintf("%s %s", prop, cont), true
}

func getHost(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return parsed.Host, nil
}

func newHostBreaker(host string, cfg breakerConfig) *gobreaker.CircuitBreaker[*OGTags] {
	st := gobreaker.Settings{
		Name:        fmt.Sprintf("%s-breaker", host),
		MaxRequests: uint32(cfg.maxRequest),
		Interval:    cfg.interval,
		Timeout:     cfg.timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= uint32(cfg.tripRequestCount) && failureRatio >= cfg.tripFailureRatio
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("Circuit breaker '%s' changed from '%s' to '%s'\n", name, from, to)
		},
	}
	return gobreaker.NewCircuitBreaker[*OGTags](st)
}
