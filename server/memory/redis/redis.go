package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"github.com/achetronic/magec/server/memory"
)

func init() {
	memory.Register(&redisProvider{})
}

type redisProvider struct{}

func (p *redisProvider) Type() string        { return "redis" }
func (p *redisProvider) DisplayName() string { return "Redis" }

func (p *redisProvider) SupportedCategories() []memory.Category {
	return []memory.Category{memory.CategorySession}
}

func (p *redisProvider) ConfigFields() []memory.FieldSpec {
	return []memory.FieldSpec{
		{Key: "connectionString", Label: "Connection String", Type: "text", Required: true, Placeholder: "redis://localhost:6379/0"},
		{Key: "ttl", Label: "TTL", Type: "text", Placeholder: "24h", Default: "24h"},
	}
}

func (p *redisProvider) Ping(ctx context.Context, config map[string]interface{}) memory.HealthResult {
	connStr, _ := config["connectionString"].(string)
	if connStr == "" {
		return memory.HealthResult{Healthy: false, Detail: "no connectionString configured"}
	}

	opts, err := goredis.ParseURL(connStr)
	if err != nil {
		return memory.HealthResult{Healthy: false, Detail: fmt.Sprintf("invalid connectionString: %s", err)}
	}

	client := goredis.NewClient(opts)
	defer client.Close()

	if err := client.Ping(ctx).Err(); err != nil {
		return memory.HealthResult{Healthy: false, Detail: fmt.Sprintf("ping failed: %s", err)}
	}
	return memory.HealthResult{Healthy: true, Detail: "connected"}
}
