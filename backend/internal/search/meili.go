// Package search integrates Meilisearch as a derived product index (plan §3.5).
// MySQL stays the source of truth; the index is kept in sync via Asynq jobs.
// A thin net/http client (no SDK) keeps dependencies light. All search reads go
// through SearchService, which falls back to the DB when Meilisearch is down,
// so search keeps working before the Meilisearch container is running.
package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mrkaak/restaurant-api/internal/config"
)

// IndexName is the Meilisearch index for products.
const IndexName = "products"

// Document is the shape stored in Meilisearch. Multi-language fields live in one
// document; the API filters/boosts by locale (plan §3.5).
type Document struct {
	ID          uint64   `json:"id"`
	Slug        string   `json:"slug"`
	NameEN      string   `json:"name_en,omitempty"`
	NameAR      string   `json:"name_ar,omitempty"`
	NameFR      string   `json:"name_fr,omitempty"`
	DescEN      string   `json:"desc_en,omitempty"`
	DescAR      string   `json:"desc_ar,omitempty"`
	DescFR      string   `json:"desc_fr,omitempty"`
	Category    string   `json:"category,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	PriceFrom   int64    `json:"price_from"`
	IsAvailable bool     `json:"is_available"`
	ImageURL    string   `json:"image_url,omitempty"`
}

// Client is a minimal Meilisearch HTTP client.
type Client struct {
	host      string
	masterKey string
	http      *http.Client
}

func NewClient(cfg config.Meili) *Client {
	return &Client{host: cfg.Host, masterKey: cfg.MasterKey, http: &http.Client{Timeout: 5 * time.Second}}
}

// EnsureIndex creates the index (if missing) and configures searchable +
// filterable attributes, typo tolerance, and synonyms. Idempotent.
func (c *Client) EnsureIndex(ctx context.Context) error {
	_, _ = c.do(ctx, http.MethodPost, "/indexes", map[string]any{
		"uid": IndexName, "primaryKey": "id",
	})
	if _, err := c.do(ctx, http.MethodPatch, "/indexes/"+IndexName+"/settings", map[string]any{
		"searchableAttributes": []string{"name_en", "name_ar", "name_fr", "desc_en", "desc_ar", "desc_fr", "category", "tags"},
		"filterableAttributes": []string{"category", "is_available", "price_from"},
		"synonyms": map[string][]string{
			"kaak":  {"kaake", "kaak"},
			"knefe": {"knefi", "kunafa", "kanafeh"},
		},
	}); err != nil {
		return err
	}
	return nil
}

// UpsertDocuments adds/updates documents in the index.
func (c *Client) UpsertDocuments(ctx context.Context, docs []Document) error {
	_, err := c.do(ctx, http.MethodPost, "/indexes/"+IndexName+"/documents", docs)
	return err
}

// DeleteDocument removes one product document.
func (c *Client) DeleteDocument(ctx context.Context, id uint64) error {
	_, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("/indexes/%s/documents/%d", IndexName, id), nil)
	return err
}

// SearchHit is a single result row returned to the caller.
type SearchHit map[string]any

// Search runs a query, optionally filtering to available items only.
func (c *Client) Search(ctx context.Context, q string, availableOnly bool, limit int) ([]SearchHit, error) {
	body := map[string]any{"q": q, "limit": limit}
	if availableOnly {
		body["filter"] = "is_available = true"
	}
	raw, err := c.do(ctx, http.MethodPost, "/indexes/"+IndexName+"/search", body)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Hits []SearchHit `json:"hits"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	return resp.Hits, nil
}

// Health reports whether Meilisearch is reachable.
func (c *Client) Health(ctx context.Context) error {
	_, err := c.do(ctx, http.MethodGet, "/health", nil)
	return err
}

func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.host+path, reader)
	if err != nil {
		return nil, err
	}
	if c.masterKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.masterKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return raw, fmt.Errorf("meili %s %s: status %d", method, path, resp.StatusCode)
	}
	return raw, nil
}
