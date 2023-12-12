package minio

import (
	"bytes"
	"context"
	pb "github.com/Verce11o/yata-protos/gen/go/tweets"
	"github.com/minio/minio-go/v7"
	"go.opentelemetry.io/otel/trace"
	"net/url"
	"time"
)

const (
	userTweetsName  = "user-tweets"
	imageExpireTime = time.Hour * 24
)

type TweetMinio struct {
	minio  *minio.Client
	tracer trace.Tracer
}

func NewTweetMinio(minio *minio.Client, tracer trace.Tracer) *TweetMinio {
	return &TweetMinio{minio: minio, tracer: tracer}
}

func (t *TweetMinio) AddTweetImage(ctx context.Context, image *pb.Image, fileName string) (string, error) {
	ctx, span := t.tracer.Start(ctx, "tweetMinio.AddImage")
	defer span.End()

	reader := bytes.NewReader(image.GetChunk())

	_, err := t.minio.PutObject(
		ctx,
		userTweetsName,
		fileName,
		reader,
		reader.Size(),
		minio.PutObjectOptions{ContentType: image.GetContentType()},
	)
	if err != nil {
		return "", err
	}

	u, err := t.minio.PresignedGetObject(ctx, userTweetsName, fileName, imageExpireTime, url.Values{})

	if err != nil {
		return "", err
	}

	return u.String(), nil
}

// GetTweetImage returns image url on minio
func (t *TweetMinio) GetTweetImage(ctx context.Context, fileName string) (string, error) {
	ctx, span := t.tracer.Start(ctx, "tweetMinio.GetImage")
	defer span.End()

	u, err := t.minio.PresignedGetObject(ctx, userTweetsName, fileName, imageExpireTime, url.Values{})

	if err != nil {
		return "", err
	}

	return u.String(), nil
}

//func (t *TweetMinio) UpdateTweetImage(ctx context.Context, oldName string, newName string) error {
//	ctx, span := t.tracer.Start(ctx, "tweetMinio.UpdateTweetImage")
//	defer span.End()
//
//	copyDestOpts := minio.CopyDestOptions{
//		Bucket: userTweetsName,
//		Object: newName,
//	}
//
//	copySrcOpts := minio.CopySrcOptions{
//		Bucket: userTweetsName,
//		Object: oldName,
//	}
//
//	if _, err := t.minio.CopyObject(ctx, copyDestOpts, copySrcOpts); err != nil {
//		return err
//	}
//
//	if err := t.minio.RemoveObject(ctx, userTweetsName, oldName, minio.RemoveObjectOptions{}); err != nil {
//		return err
//	}
//
//	return nil
//}

func (t *TweetMinio) DeleteFile(ctx context.Context, fileName string) error {
	ctx, span := t.tracer.Start(ctx, "tweetMinio.DeleteFile")
	defer span.End()

	if err := t.minio.RemoveObject(ctx, userTweetsName, fileName, minio.RemoveObjectOptions{}); err != nil {
		return err
	}

	return nil
}
