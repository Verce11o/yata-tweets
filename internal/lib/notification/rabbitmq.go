package notification

import (
	"context"
)

type TweetPublisher interface {
	Publish(ctx context.Context, message []byte) error
}
