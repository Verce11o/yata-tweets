package repository

import (
	"context"
	pb "github.com/Verce11o/yata-protos/gen/go/tweets"
	"github.com/Verce11o/yata-tweets/internal/domain"
)

type RedisRepository interface { // maybe refactor ?
	GetTweetByIDCtx(ctx context.Context, key string) (*domain.Tweet, error)
	SetByIDCtx(ctx context.Context, tweetID string, tweet *domain.Tweet) error
	DeleteTweetByIDCtx(ctx context.Context, tweetID string) error
}

type PostgresRepository interface {
	CreateTweet(ctx context.Context, input *pb.CreateTweetRequest, imageName string) (string, error)
	GetTweet(ctx context.Context, tweetID string) (*domain.Tweet, error)
	UpdateTweet(ctx context.Context, input *pb.UpdateTweetRequest, imageName string) (*domain.Tweet, error)
	DeleteTweet(ctx context.Context, tweetID string) error
}

type MinioRepository interface {
	AddTweetImage(ctx context.Context, image *pb.Image, fileName string) error
	GetTweetImage(ctx context.Context, fileName string) ([]byte, string, error)
	UpdateTweetImage(ctx context.Context, oldName string, newName string) error
	DeleteFile(ctx context.Context, fileName string) error
}