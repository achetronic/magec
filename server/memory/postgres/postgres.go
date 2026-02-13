package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/achetronic/magec/server/memory"
)

func init() {
	memory.Register(&postgresProvider{})
}

type postgresProvider struct{}

func (p *postgresProvider) Type() string        { return "postgres" }
func (p *postgresProvider) DisplayName() string { return "PostgreSQL" }

func (p *postgresProvider) SupportedCategories() []memory.Category {
	return []memory.Category{memory.CategoryLongTerm}
}

func (p *postgresProvider) ConfigSchema() memory.Schema {
	return memory.Schema{
		"type": "object",
		"properties": memory.Schema{
			"connectionString": memory.Schema{
				"type":          "string",
				"title":         "Connection String",
				"minLength":     1,
				"x-placeholder": "postgres://user:pass@localhost:5432/db?sslmode=disable",
			},
		},
		"required": []string{"connectionString"},
	}
}

func (p *postgresProvider) Ping(ctx context.Context, config map[string]interface{}) memory.HealthResult {
	connStr, _ := config["connectionString"].(string)
	if connStr == "" {
		return memory.HealthResult{Healthy: false, Detail: "no connection string configured"}
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return memory.HealthResult{Healthy: false, Detail: fmt.Sprintf("open failed: %s", err)}
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return memory.HealthResult{Healthy: false, Detail: fmt.Sprintf("ping failed: %s", err)}
	}
	return memory.HealthResult{Healthy: true, Detail: "connected"}
}
