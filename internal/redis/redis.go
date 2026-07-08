package redis

import (
	"github.com/aipass/aipass/internal/config"
	goredis "github.com/redis/go-redis/v9"
)

func Connect(cfg config.RedisConfig) *goredis.Client {
	return goredis.NewClient(&goredis.Options{Addr: cfg.Addr, Password: cfg.Password, DB: cfg.DB})
}
