package elasticsearch

import (
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

type IndexTemplate struct {
	client *elasticsearch.Client
}

func NewIndexTemplate(client *elasticsearch.Client) *IndexTemplate {
	return &IndexTemplate{client: client}
}

func (it *IndexTemplate) CreateCategoryTemplate() error {
	template := `{
        "index_patterns": ["*-digital-discovery-categories-*"],
        "template": {
            "settings": {
                "number_of_shards": 3,
                "number_of_replicas": 1,
                "refresh_interval": "1s",
                "analysis": {
                    "analyzer": {
                        "custom_analyzer": {
                            "type": "custom",
                            "tokenizer": "standard",
                            "filter": ["lowercase", "asciifolding"]
                        }
                    }
                }
            },
            "mappings": {
                "properties": {
                    "id": {
                        "type": "keyword"
                    },
                    "name": {
                        "type": "text",
                        "analyzer": "custom_analyzer",
                        "fields": {
                            "keyword": {
                                "type": "keyword",
                                "ignore_above": 256
                            }
                        }
                    },
                    "description": {
                        "type": "text",
                        "analyzer": "custom_analyzer"
                    },
                    "created_at": {
                        "type": "date"
                    },
                    "updated_at": {
                        "type": "date"
                    },
                    "version": {
                        "type": "long"
                    },
                    "sync_status": {
                        "type": "keyword"
                    },
                    "last_sync": {
                        "type": "date"
                    }
                }
            }
        },
        "priority": 100,
        "version": 1,
        "_meta": {
            "description": "Template for category indices",
            "service": "digital-discovery"
        }
    }`

	resp, err := it.client.Indices.PutIndexTemplate(
		"categories-template",
		strings.NewReader(template),
	)
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("error creating template: %s", resp.String())
	}
	return nil
}
