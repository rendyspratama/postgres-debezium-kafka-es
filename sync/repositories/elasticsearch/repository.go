package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

// ErrInvalidConfig represents a configuration error
var ErrInvalidConfig = fmt.Errorf("invalid elasticsearch configuration")

// Config holds Elasticsearch client configuration
type Config struct {
	Addresses      []string
	Username       string
	Password       string
	MaxRetries     int
	RetryBackoff   time.Duration
	EnableRetry    bool
	MaxConns       int
	RequestTimeout time.Duration
	GzipEnabled    bool
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if len(c.Addresses) == 0 {
		return fmt.Errorf("%w: addresses cannot be empty", ErrInvalidConfig)
	}
	if c.MaxRetries < 0 {
		return fmt.Errorf("%w: max retries cannot be negative", ErrInvalidConfig)
	}
	if c.MaxConns <= 0 {
		c.MaxConns = 100 // default value
	}
	if c.RequestTimeout == 0 {
		c.RequestTimeout = 30 * time.Second // default timeout
	}
	return nil
}

// Repository defines the interface for Elasticsearch operations
type Repository interface {
	// Index operations
	Index(ctx context.Context, index, id string, body io.Reader) error
	Update(ctx context.Context, index, id string, body io.Reader) error
	Delete(ctx context.Context, index, id string) error
	Search(ctx context.Context, index string, query interface{}) ([]json.RawMessage, error)
	Bulk(ctx context.Context, body io.Reader) error
	Ping(ctx context.Context) error
	IndexExists(ctx context.Context, index string) (bool, error)

	// Setup and maintenance
	CheckHealth(ctx context.Context) error
	CreateTemplate(ctx context.Context) error
	CreateLifecyclePolicy(ctx context.Context, name string) error
	VerifySetup(ctx context.Context) error

	// Cleanup
	Close() error
}

// Operation represents a bulk operation
type Operation struct {
	Action string
	Index  string
	ID     string
	Body   interface{}
}

// esRepository implements the Repository interface
type esRepository struct {
	client *elasticsearch.Client
	config *Config
}

// NewRepository creates a new Elasticsearch repository
func NewRepository(cfg *Config) (Repository, error) {
	if cfg == nil {
		return nil, fmt.Errorf("%w: config cannot be nil", ErrInvalidConfig)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	transport := &http.Transport{
		MaxIdleConnsPerHost: cfg.MaxConns,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	esCfg := elasticsearch.Config{
		Addresses:    cfg.Addresses,
		Username:     cfg.Username,
		Password:     cfg.Password,
		MaxRetries:   cfg.MaxRetries,
		RetryBackoff: func(i int) time.Duration { return cfg.RetryBackoff },
		Transport:    transport,
	}

	if cfg.GzipEnabled {
		esCfg.Header = http.Header{
			"Accept-Encoding": []string{"gzip"},
		}
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	repo := &esRepository{
		client: client,
		config: cfg,
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := repo.CheckHealth(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to elasticsearch: %w", err)
	}

	return repo, nil
}

func (r *esRepository) Index(ctx context.Context, index, id string, body io.Reader) error {
	if index == "" || id == "" {
		return fmt.Errorf("index and id cannot be empty")
	}

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: id,
		Body:       body,
		Refresh:    "true",
		Timeout:    r.config.RequestTimeout,
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to execute index request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("index error: status=%s body=%s", res.Status(), string(bodyBytes))
	}
	return nil
}

func (r *esRepository) Update(ctx context.Context, index, id string, body io.Reader) error {
	req := esapi.UpdateRequest{
		Index:      index,
		DocumentID: id,
		Body:       body,
		Timeout:    r.config.RequestTimeout,
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to execute update request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("update error: %s", res.String())
	}
	return nil
}

func (r *esRepository) Delete(ctx context.Context, index, id string) error {
	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: id,
		Timeout:    r.config.RequestTimeout,
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to execute delete request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("delete error: %s", res.String())
	}
	return nil
}

func (r *esRepository) Bulk(ctx context.Context, body io.Reader) error {
	req := esapi.BulkRequest{
		Body:    body,
		Refresh: "true",
		Timeout: r.config.RequestTimeout,
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to execute bulk request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk error: %s", res.String())
	}
	return nil
}

func (r *esRepository) CheckHealth(ctx context.Context) error {
	res, err := r.client.Cluster.Health(
		r.client.Cluster.Health.WithContext(ctx),
		r.client.Cluster.Health.WithTimeout(r.config.RequestTimeout),
	)
	if err != nil {
		return fmt.Errorf("failed to check cluster health: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("health check error: %s", res.String())
	}
	return nil
}

func (r *esRepository) CreateTemplate(ctx context.Context) error {
	template := map[string]interface{}{
		"index_patterns": []string{"development-digital-discovery-categories-*"},
		"priority":       500, // Add high priority to avoid conflicts
		"template": map[string]interface{}{
			"settings": map[string]interface{}{
				"number_of_shards":   1,
				"number_of_replicas": 1,
			},
			"mappings": map[string]interface{}{
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type": "keyword",
					},
					"name": map[string]interface{}{
						"type": "text",
						"fields": map[string]interface{}{
							"keyword": map[string]interface{}{
								"type":         "keyword",
								"ignore_above": 256,
							},
						},
					},
					"description": map[string]interface{}{
						"type": "text",
					},
					"status": map[string]interface{}{
						"type": "keyword",
					},
					"sync_status": map[string]interface{}{
						"type": "keyword",
					},
					"last_sync": map[string]interface{}{
						"type": "date",
					},
					"created_at": map[string]interface{}{
						"type": "date",
					},
					"updated_at": map[string]interface{}{
						"type": "date",
					},
				},
			},
		},
		// Add metadata
		"version": 1,
		"_meta": map[string]interface{}{
			"description": "Template for digital discovery categories",
			"application": "digital-discovery",
		},
	}

	// Delete existing template if it exists
	deleteRes, err := r.client.Indices.DeleteIndexTemplate(
		"categories-template",
		r.client.Indices.DeleteIndexTemplate.WithContext(ctx),
	)
	if err != nil && !strings.Contains(err.Error(), "404") {
		return fmt.Errorf("failed to delete existing template: %w", err)
	}
	if deleteRes != nil {
		defer deleteRes.Body.Close()
	}

	// Create new template
	res, err := r.client.Indices.PutIndexTemplate(
		"categories-template",
		esutil.NewJSONReader(template),
		r.client.Indices.PutIndexTemplate.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("template creation failed: status=%s body=%s", res.Status(), body)
	}

	// Create initial index
	initialIndex := fmt.Sprintf("development-digital-discovery-categories-%s", time.Now().Format("2006-01"))
	if err := r.createInitialIndex(ctx, initialIndex); err != nil {
		return fmt.Errorf("failed to create initial index: %w", err)
	}

	// Create alias
	if err := r.createAlias(ctx, initialIndex); err != nil {
		return fmt.Errorf("failed to create alias: %w", err)
	}

	return nil
}

// Helper function to create initial index
func (r *esRepository) createInitialIndex(ctx context.Context, indexName string) error {
	createRes, err := r.client.Indices.Create(
		indexName,
		r.client.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer createRes.Body.Close()

	// If index already exists (400 error), that's fine
	if createRes.IsError() && createRes.StatusCode != 400 {
		body, _ := io.ReadAll(createRes.Body)
		return fmt.Errorf("index creation failed: status=%s body=%s", createRes.Status(), body)
	}

	// Wait for index to be ready
	time.Sleep(2 * time.Second)
	return nil
}

// Helper function to create alias
func (r *esRepository) createAlias(ctx context.Context, indexName string) error {
	aliasBody := map[string]interface{}{
		"actions": []map[string]interface{}{
			{
				"add": map[string]interface{}{
					"index": indexName,
					"alias": "digital-discovery-categories",
				},
			},
		},
	}

	aliasRes, err := r.client.Indices.UpdateAliases(
		esutil.NewJSONReader(aliasBody),
		r.client.Indices.UpdateAliases.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer aliasRes.Body.Close()

	if aliasRes.IsError() {
		body, _ := io.ReadAll(aliasRes.Body)
		return fmt.Errorf("alias creation failed: status=%s body=%s", aliasRes.Status(), body)
	}

	return nil
}

func (r *esRepository) CreateLifecyclePolicy(ctx context.Context, name string) error {
	// First check if policy exists
	existsRes, err := r.client.ILM.GetLifecycle(
		r.client.ILM.GetLifecycle.WithPolicy(name),
		r.client.ILM.GetLifecycle.WithContext(ctx),
	)
	if err == nil && !existsRes.IsError() {
		// Policy already exists
		return nil
	}

	policy := map[string]interface{}{
		"policy": map[string]interface{}{
			"phases": map[string]interface{}{
				"hot": map[string]interface{}{
					"actions": map[string]interface{}{
						"rollover": map[string]interface{}{
							"max_size": "50gb",
							"max_age":  "30d",
						},
					},
				},
			},
		},
	}

	policyBytes, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}

	res, err := r.client.ILM.PutLifecycle(
		name,
		r.client.ILM.PutLifecycle.WithBody(bytes.NewReader(policyBytes)),
		r.client.ILM.PutLifecycle.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to create lifecycle policy: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("lifecycle policy creation failed: status=%s body=%s", res.Status(), body)
	}

	return nil
}

func (r *esRepository) VerifySetup(ctx context.Context) error {
	// Check cluster health
	healthRes, err := r.client.Cluster.Health(
		r.client.Cluster.Health.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to check cluster health: %w", err)
	}
	defer healthRes.Body.Close()

	if healthRes.IsError() {
		return fmt.Errorf("cluster is not healthy: %s", healthRes.Status())
	}

	// Check template
	templateRes, err := r.client.Indices.GetIndexTemplate(
		r.client.Indices.GetIndexTemplate.WithName("categories-template"),
		r.client.Indices.GetIndexTemplate.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to verify template: %w", err)
	}
	defer templateRes.Body.Close()

	if templateRes.IsError() {
		return fmt.Errorf("template verification failed: %s", templateRes.Status())
	}

	// Check if current month's index exists
	currentMonth := time.Now().Format("2006-01")
	currentIndex := fmt.Sprintf("development-digital-discovery-categories-%s", currentMonth)

	// Try to create the index if it doesn't exist
	createRes, err := r.client.Indices.Create(
		currentIndex,
		r.client.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer createRes.Body.Close()

	// If index already exists (400 error), that's fine
	if createRes.IsError() && createRes.StatusCode != 400 {
		body, _ := io.ReadAll(createRes.Body)
		return fmt.Errorf("index creation failed: status=%s body=%s", createRes.Status(), body)
	}

	// Wait for index to be ready
	time.Sleep(2 * time.Second)

	// Check if alias exists
	aliasRes, err := r.client.Indices.GetAlias(
		r.client.Indices.GetAlias.WithName("digital-discovery-categories"),
		r.client.Indices.GetAlias.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to verify alias: %w", err)
	}
	defer aliasRes.Body.Close()

	if aliasRes.IsError() {
		// Try to create the alias if it doesn't exist
		aliasBody := map[string]interface{}{
			"actions": []map[string]interface{}{
				{
					"add": map[string]interface{}{
						"index": currentIndex,
						"alias": "digital-discovery-categories",
					},
				},
			},
		}

		aliasCreateRes, err := r.client.Indices.UpdateAliases(
			esutil.NewJSONReader(aliasBody),
			r.client.Indices.UpdateAliases.WithContext(ctx),
		)
		if err != nil {
			return fmt.Errorf("failed to create alias: %w", err)
		}
		defer aliasCreateRes.Body.Close()

		if aliasCreateRes.IsError() {
			body, _ := io.ReadAll(aliasCreateRes.Body)
			return fmt.Errorf("alias creation failed: status=%s body=%s", aliasCreateRes.Status(), body)
		}
	}

	return nil
}

func (r *esRepository) Close() error {
	// No need to close the transport as it's managed by the ES client
	return nil
}

// Search executes a search query in Elasticsearch
func (r *esRepository) Search(ctx context.Context, index string, query interface{}) ([]json.RawMessage, error) {
	// Convert query to JSON
	queryBody, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := esapi.SearchRequest{
		Index:   []string{index},
		Body:    bytes.NewReader(queryBody),
		Timeout: r.config.RequestTimeout,
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}

	// Parse response
	var result struct {
		Hits struct {
			Hits []struct {
				Source json.RawMessage `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	// Extract source documents
	var docs []json.RawMessage
	for _, hit := range result.Hits.Hits {
		docs = append(docs, hit.Source)
	}

	return docs, nil
}

func (r *esRepository) Ping(ctx context.Context) error {
	res, err := r.client.Ping(
		r.client.Ping.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to ping elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch ping failed: %s", res.String())
	}
	return nil
}

func (r *esRepository) IndexExists(ctx context.Context, index string) (bool, error) {
	res, err := r.client.Indices.Exists([]string{index})
	if err != nil {
		return false, err
	}
	return res.StatusCode != 404, nil
}
