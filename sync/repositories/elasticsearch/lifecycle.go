package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type LifecyclePolicy struct {
	client *elasticsearch.Client
}

func NewLifecyclePolicy(client *elasticsearch.Client) *LifecyclePolicy {
	return &LifecyclePolicy{client: client}
}

func (lp *LifecyclePolicy) validatePolicy() error {
	if lp.client == nil {
		return fmt.Errorf("elasticsearch client is nil")
	}
	return nil
}

func (lp *LifecyclePolicy) CreatePolicy(ctx context.Context) error {
	if err := lp.validatePolicy(); err != nil {
		return fmt.Errorf("policy validation failed: %w", err)
	}

	policy := `{
		"policy": {
			"phases": {
				"hot": {
					"min_age": "0ms",
					"actions": {
						"rollover": {
							"max_age": "30d",
							"max_size": "50gb"
						},
						"set_priority": {
							"priority": 100
						}
					}
				},
				"warm": {
					"min_age": "30d",
					"actions": {
						"shrink": {
							"number_of_shards": 1
						},
						"forcemerge": {
							"max_num_segments": 1
						},
						"set_priority": {
							"priority": 50
						}
					}
				},
				"cold": {
					"min_age": "60d",
					"actions": {
						"set_priority": {
							"priority": 0
						}
					}
				},
				"delete": {
					"min_age": "90d",
					"actions": {
						"delete": {}
					}
				}
			}
		}
	}`

	req := esapi.ILMPutLifecycleRequest{
		Policy: "digital-discovery-policy",
		Body:   strings.NewReader(policy),
	}

	res, err := req.Do(ctx, lp.client)
	if err != nil {
		return fmt.Errorf("failed to create lifecycle policy: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error creating lifecycle policy: %s", res.String())
	}

	return nil
}
