package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"receiver/internal/storage"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const receiverObj = "receiver"

var ttl time.Duration = 5 * time.Minute

// GetFromRedis return pointer to TextRequest
//
// return nil if redis error or key not presented
func GetFromRedis(rdb *redis.Client, id uuid.UUID) *storage.TextRequest {
	var result storage.TextRequest
	val, err := rdb.Get(context.Background(), getRedisKey(id.String())).Result()
	if err != nil {
		if err != redis.Nil {
			log.Error().Str("event", "get from redis").Str("requestID", id.String()).Err(err).Msg("Redis error")
		}
		return nil
	} else {
		err = json.Unmarshal([]byte(val), &result)
		if err != nil {
			log.Error().Str("event", "unmarshall redis value").Str("requestID", id.String()).Err(err).Msg("Failed to unmarshal")
			return nil
		}
	}
	return &result
}

// SetInRedis set TextRequest on key id
func SetInRedis(result *storage.TextRequest, rdb *redis.Client, id uuid.UUID) {
	redisVal, err := json.Marshal(result)
	if err != nil {
		log.Error().Str("event", "marshall value").Any("obj", *result).Err(err).Msg("Failed to marshal object")
		return
	}
	err = rdb.Set(context.Background(), getRedisKey(id.String()), string(redisVal), ttl).Err()
	if err != nil {
		log.Error().Str("event", "redis set").Any("obj", *result).Err(err).Msg("Failed to marshal object")
		return
	}
}

// getRedisKey return {obj}:{id}
func getRedisKey(id string) string {
	key := fmt.Sprintf("%s:%s", receiverObj, id)
	return key
}
