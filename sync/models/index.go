package models

import (
	"fmt"
	"time"
)

type IndexNaming struct {
	// Base pattern: {env}-{service}-{entity}-{yyyy-MM}
	// Example: prod-digital-discovery-categories-2024-03

	Environment string    // prod, stg, dev
	Service     string    // digital-discovery
	Entity      string    // categories, products, etc.
	Date        time.Time // For time-based rotation
}

func (in *IndexNaming) GetIndexName() string {
	return fmt.Sprintf("%s-%s-%s-%s",
		in.Environment,
		in.Service,
		in.Entity,
		in.Date.Format("2006-01"))
}

func (in *IndexNaming) GetAliasName() string {
	return fmt.Sprintf("%s-%s-%s",
		in.Environment,
		in.Service,
		in.Entity)
}
