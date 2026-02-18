package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strconv"
	"time"

	"github.com/Critma/textAnalyzer/analyzer/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const analyzerObj = "analyze"

var ttl time.Duration = 10 * time.Minute

// GetFromRedis return pointer to JsonAnalyze
//
// return nil if redis error or key(text) not presented
func GetFromRedis(rdb *redis.Client, text string) *models.JsonAnalyze {
	var result models.JsonAnalyze
	val, err := rdb.Get(context.Background(), getRedisKey(text)).Result()
	if err != nil {
		if err != redis.Nil {
			log.Error().Str("event", "get from redis").Str("text", text).Err(err).Msg("cache not found")
		}
		return nil
	} else {
		err = json.Unmarshal([]byte(val), &result)
		if err != nil {
			log.Error().Str("event", "unmarshall redis value").Str("toUnmarshall", val).Err(err).Msg("Failed to unmarshal")
			return nil
		}
	}
	return &result
}

// SetInRedis set JsonAnalyze on key (text)
func SetInRedis(result *models.JsonAnalyze, rdb *redis.Client, text string) {
	redisVal, err := json.Marshal(result)
	if err != nil {
		log.Error().Str("event", "marshall value").Any("obj", *result).Err(err).Msg("Failed to marshal object")
		return
	}
	err = rdb.Set(context.Background(), getRedisKey(text), string(redisVal), ttl).Err()
	if err != nil {
		log.Error().Str("event", "redis set").Any("obj", *result).Err(err).Msg("Failed to marshal object")
		return
	}
}

// getRedisKey hash the text with fnv
//
// return obj:{hash}
func getRedisKey(text string) string {
	h := fnv.New32a()
	h.Write([]byte(text))
	hashValue := h.Sum32()

	key := fmt.Sprintf("%s:%s", analyzerObj, strconv.FormatUint(uint64(hashValue), 10))
	return key
}
