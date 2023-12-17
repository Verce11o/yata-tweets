package service

import (
	"context"
	pb "github.com/Verce11o/yata-protos/gen/go/tweets"
	"github.com/Verce11o/yata-tweets/internal/domain"
)

type Tweet interface {
	CreateTweet(ctx context.Context, input *pb.CreateTweetRequest) (string, error)
	GetTweet(ctx context.Context, tweetID string) (domain.Tweet, error)
	GetAllTweets(ctx context.Context, input *pb.GetAllTweetsRequest) ([]*pb.Tweet, string, error)
	UpdateTweet(ctx context.Context, input *pb.UpdateTweetRequest) (*domain.Tweet, error)
	DeleteTweet(ctx context.Context, input *pb.DeleteTweetRequest) error
}
