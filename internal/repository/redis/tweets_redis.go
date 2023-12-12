package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Verce11o/yata-tweets/internal/domain"
	"github.com/Verce11o/yata-tweets/internal/lib/grpc_errors"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
	"time"
)

const (
	tweetTTL = 3600
)

type TweetsRedis struct {
	client *redis.Client
	tracer trace.Tracer
}

func NewTweetsRedis(client *redis.Client, tracer trace.Tracer) *TweetsRedis {
	return &TweetsRedis{client: client, tracer: tracer}
}

func (r *TweetsRedis) GetTweetByIDCtx(ctx context.Context, tweetID string) (*domain.Tweet, error) {
	ctx, span := r.tracer.Start(ctx, "tweetRedis.GetTweetByIDCtx")
	defer span.End()

	tweetBytes, err := r.client.Get(ctx, r.createKey(tweetID)).Bytes()

	if err != nil {
		if !errors.Is(err, redis.Nil) {
			return nil, grpc_errors.ErrNotFound
		}
		return nil, err
	}

	var tweet domain.Tweet

	if err = json.Unmarshal(tweetBytes, &tweet); err != nil {
		return nil, err
	}

	return &tweet, nil
}

func (r *TweetsRedis) SetByIDCtx(ctx context.Context, tweetID string, tweet *domain.Tweet) error {
	ctx, span := r.tracer.Start(ctx, "tweetRedis.SetByIDCtx")
	defer span.End()

	tweetBytes, err := json.Marshal(tweet)

	if err != nil {
		return err
	}

	return r.client.Set(ctx, r.createKey(tweetID), tweetBytes, time.Second*time.Duration(tweetTTL)).Err()
}

func (r *TweetsRedis) DeleteTweetByIDCtx(ctx context.Context, tweetID string) error {
	ctx, span := r.tracer.Start(ctx, "tweetRedis.DeleteTweetByIDCtx")
	defer span.End()

	return r.client.Del(ctx, r.createKey(tweetID)).Err()
}

func (r *TweetsRedis) createKey(key string) string {
	return fmt.Sprintf("tweet:%s", key)
}
