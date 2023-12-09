package service

import (
	"context"
	pb "github.com/Verce11o/yata-protos/gen/go/tweets"
	"github.com/Verce11o/yata-tweets/internal/domain"
)

type Tweet interface {
	CreateTweet(ctx context.Context, input *pb.CreateTweetRequest) (string, error)
	GetTweet(ctx context.Context, tweetID string) (domain.Tweet, error)
	GetTweetImage(ctx context.Context, imageName string) (*pb.Image, error)
	UpdateTweet(ctx context.Context, input *pb.UpdateTweetRequest) (*domain.Tweet, error)
	DeleteTweet(ctx context.Context, input *pb.DeleteTweetRequest) error
}
