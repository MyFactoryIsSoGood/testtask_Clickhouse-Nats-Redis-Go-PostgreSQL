package cache

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"hezzl/models"
	"os"
	"strconv"
	"time"
)

var ctx = context.Background()
var rdb *redis.Client

func Connect() error {
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	var addr string
	var pass string

	if os.Getenv("APP_ENV") == "docker" {
		addr = os.Getenv("REDIS_HOST")
		pass = os.Getenv("REDIS_PASS")
	} else {
		addr = "localhost"
		pass = "123qwe123"
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       db,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return err
	}
	return nil
}

func GetItems(key string) (bool, []models.Item) {
	jsonData, _ := rdb.Get(ctx, key).Result()
	if jsonData == "" {
		return false, nil
	}

	var resp []models.Item
	_ = json.Unmarshal([]byte(jsonData), &resp)

	return true, resp
}

func SetItems(key string, items []models.Item) {
	jsonData, _ := json.Marshal(items)
	rdb.Set(ctx, key, jsonData, time.Minute)
}

func InvalidateItems() error {
	err := rdb.Del(ctx, "items").Err()
	return err
}
