package grpc

import (
	"context"
	pb "github.com/Verce11o/yata-protos/gen/go/tweets"
	"github.com/Verce11o/yata-tweets/internal/lib/grpc_errors"
	"github.com/Verce11o/yata-tweets/internal/service"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/status"

	"go.uber.org/zap"
)

type TweetGRPC struct {
	log     *zap.SugaredLogger
	tracer  trace.Tracer
	service service.Tweet
	pb.UnimplementedTweetsServer
}

func NewTweetGRPC(log *zap.SugaredLogger, tracer trace.Tracer, service service.Tweet) *TweetGRPC {
	return &TweetGRPC{log: log, tracer: tracer, service: service}
}

func (t *TweetGRPC) CreateTweet(ctx context.Context, input *pb.CreateTweetRequest) (*pb.CreateTweetResponse, error) {
	t.log.Info(ctx.Value)
	ctx, span := t.tracer.Start(ctx, "CreateTweet")
	defer span.End()

	tweetID, err := t.service.CreateTweet(ctx, input)

	if err != nil {
		t.log.Errorf("CreateTweet: %v", err.Error())
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "CreateTweet: %v", err)
	}

	return &pb.CreateTweetResponse{TweetId: tweetID}, nil
}

func (t *TweetGRPC) GetTweet(ctx context.Context, input *pb.GetTweetRequest) (*pb.Tweet, error) {
	ctx, span := t.tracer.Start(ctx, "GRPC.GetTweet")
	defer span.End()

	tweet, err := t.service.GetTweet(ctx, input.GetTweetId())

	if err != nil {
		t.log.Errorf("GetTweet: %v", err.Error())
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "GetTweet: %v", err)
	}

	tweetImage := &pb.Image{}

	if tweet.ImageName != "" {
		tweetImage, err = t.service.GetTweetImage(ctx, tweet.ImageName)
		if err != nil {
			t.log.Errorf("GetTweet: %v", err.Error())
		}
	}
	return &pb.Tweet{
		UserId:  tweet.UserID.String(),
		TweetId: tweet.TweetID.String(),
		Text:    tweet.Text,
		Image:   tweetImage,
	}, nil

}

func (t *TweetGRPC) UpdateTweet(ctx context.Context, input *pb.UpdateTweetRequest) (*pb.Tweet, error) {
	ctx, span := t.tracer.Start(ctx, "UpdateTweet")
	defer span.End()

	tweet, err := t.service.UpdateTweet(ctx, input)

	if err != nil {
		t.log.Errorf("UpdateTweet: %v", err.Error())
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "UpdateTweet: %v", err)
	}

	tweetImage := &pb.Image{}
	if tweet.ImageName != "" {
		tweetImage, err = t.service.GetTweetImage(ctx, tweet.ImageName)

		if err != nil {
			t.log.Errorf("GetTweet: %v", err.Error())
		}
	}

	return &pb.Tweet{
		UserId:  tweet.UserID.String(),
		TweetId: tweet.TweetID.String(),
		Text:    tweet.Text,
		Image:   tweetImage,
	}, nil

}
func (t *TweetGRPC) DeleteTweet(ctx context.Context, input *pb.DeleteTweetRequest) (*pb.DeleteTweetResponse, error) {
	ctx, span := t.tracer.Start(ctx, "DeleteTweet")
	defer span.End()

	err := t.service.DeleteTweet(ctx, input)

	if err != nil {
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "DeleteTweet: %v", err)
	}

	return &pb.DeleteTweetResponse{}, nil
}
