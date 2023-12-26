package service

import (
	"context"
	"encoding/json"
	"fmt"
	pb "github.com/Verce11o/yata-protos/gen/go/tweets"
	"github.com/Verce11o/yata-tweets/internal/domain"
	"github.com/Verce11o/yata-tweets/internal/lib/grpc_errors"
	"github.com/Verce11o/yata-tweets/internal/lib/notification"
	"github.com/Verce11o/yata-tweets/internal/repository"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type TweetService struct {
	log            *zap.SugaredLogger
	tracer         trace.Tracer
	tweetPublisher notification.TweetPublisher
	repo           repository.PostgresRepository
	redis          repository.RedisRepository
	minio          repository.MinioRepository
}

func NewTweetService(log *zap.SugaredLogger, tracer trace.Tracer, tweetPublisher notification.TweetPublisher, repo repository.PostgresRepository, redis repository.RedisRepository, minio repository.MinioRepository) *TweetService {
	return &TweetService{log: log, tracer: tracer, tweetPublisher: tweetPublisher, repo: repo, redis: redis, minio: minio}
}

func (t *TweetService) CreateTweet(ctx context.Context, input *pb.CreateTweetRequest) (string, error) {
	ctx, span := t.tracer.Start(ctx, "tweetService.CreateTweet")
	defer span.End()

	image := input.GetImage()

	var err error

	if image != nil {

		err = t.minio.AddTweetImage(ctx, image, image.GetName())

		if err != nil {
			t.log.Errorf("cannot add image to tweet in minio: %v", err.Error())
		}

	}

	tweetID, err := t.repo.CreateTweet(ctx, input, image.GetName())

	if err != nil {
		return "", err
	}

	SendNewTweetNotification := domain.SendNewTweetNotification{
		FromUserID: input.GetUserId(),
		ShortTitle: fmt.Sprintf("%v...", input.GetText()[:8]),
	}

	messageBytes, err := json.Marshal(SendNewTweetNotification)

	if err != nil {
		return "", err
	}

	err = t.tweetPublisher.Publish(ctx, messageBytes)

	if err != nil {
		return "", err
	}

	return tweetID, nil
}

func (t *TweetService) GetTweet(ctx context.Context, tweetID string) (domain.Tweet, error) {
	ctx, span := t.tracer.Start(ctx, "tweetService.GetTweet")
	defer span.End()

	cachedTweet, err := t.redis.GetTweetByIDCtx(ctx, tweetID)

	if err != nil {
		t.log.Infof("cannot get tweet by id in redis: %v", err.Error())
	}

	if cachedTweet != nil {
		t.log.Info("returned cache")
		return *cachedTweet, nil
	}

	tweet, err := t.repo.GetTweet(ctx, tweetID)

	if err != nil {
		t.log.Errorf("cannot get tweet by id in postgres: %v", err.Error())
		return domain.Tweet{}, err
	}

	if err := t.redis.SetByIDCtx(ctx, tweetID, tweet); err != nil {
		t.log.Errorf("cannot set tweet by id in redis: %v", err.Error())
	}

	return *tweet, nil

}

func (t *TweetService) GetAllTweets(ctx context.Context, input *pb.GetAllTweetsRequest) ([]*pb.Tweet, string, error) {
	ctx, span := t.tracer.Start(ctx, "tweetService.GetAllTweets")
	defer span.End()

	tweets, nextCursor, err := t.repo.GetAllTweets(ctx, input.GetCursor())

	if err != nil {
		t.log.Errorf("cannot get all tweets by cursor: %v err: %v", input.GetCursor(), err)
	}

	return tweets, nextCursor, nil

}

func (t *TweetService) UpdateTweet(ctx context.Context, input *pb.UpdateTweetRequest) (*domain.Tweet, error) {
	ctx, span := t.tracer.Start(ctx, "tweetService.UpdateTweet")
	defer span.End()

	tweet, err := t.repo.GetTweet(ctx, input.GetTweetId())

	if err != nil {
		t.log.Errorf("cannot get tweet by id in postgres: %v", err.Error())
		return nil, err
	}

	if tweet.UserID.String() != input.GetUserId() {
		t.log.Errorf("cannot update tweet by id: permission denied")
		return nil, grpc_errors.ErrPermissionDenied
	}

	image := input.GetImage()
	newImageName := tweet.ImageName

	if image != nil { // if input image is not nil, we need to update it

		err = t.minio.UpdateTweetImage(ctx, tweet.ImageName, image.GetName(), image)

		if err != nil {
			t.log.Errorf("cannot update comment image: %v", err.Error())
			return nil, err
		}

		newImageName = image.GetName()

	}

	newTweet, err := t.repo.UpdateTweet(ctx, input, newImageName)

	if err != nil {
		t.log.Errorf("cannot update tweet: %v", err.Error())
		return nil, err
	}

	if err := t.redis.DeleteTweetByIDCtx(ctx, tweet.TweetID.String()); err != nil {
		t.log.Errorf("cannot remove tweet by id in redis: %v", err.Error())
	}

	return newTweet, nil
}

func (t *TweetService) DeleteTweet(ctx context.Context, input *pb.DeleteTweetRequest) error {
	ctx, span := t.tracer.Start(ctx, "tweetService.DeleteTweet")
	defer span.End()

	tweet, err := t.repo.GetTweet(ctx, input.GetTweetId())

	if err != nil {
		t.log.Errorf("cannot get tweet by id in postgres: %v", err.Error())
		return err
	}

	if tweet.UserID.String() != input.GetUserId() {
		t.log.Errorf("cannot delete tweet by id: permission denied")
		return grpc_errors.ErrPermissionDenied
	}

	err = t.repo.DeleteTweet(ctx, tweet.TweetID.String())

	if err != nil {
		t.log.Errorf("cannot delete tweet by id: %v", err.Error())
		return err
	}

	if err := t.redis.DeleteTweetByIDCtx(ctx, tweet.TweetID.String()); err != nil {
		t.log.Errorf("cannot delete tweet by id in redis: ", err.Error())
	}

	return nil

}
