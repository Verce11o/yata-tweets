package minio

import (
	"bytes"
	"context"
	pb "github.com/Verce11o/yata-protos/gen/go/tweets"
	"github.com/minio/minio-go/v7"
	"io"
)

const (
	userTweetsName = "user-tweets"
)

type TweetMinio struct {
	minio *minio.Client
}

func NewTweetMinio(minio *minio.Client) *TweetMinio {
	return &TweetMinio{minio: minio}
}

func (t *TweetMinio) AddTweetImage(ctx context.Context, image *pb.Image, fileName string) error {

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
		return err
	}

	return nil
}

func (t *TweetMinio) GetTweetImage(ctx context.Context, fileName string) ([]byte, string, error) {
	file, err := t.minio.GetObject(ctx, userTweetsName, fileName, minio.GetObjectOptions{})

	if err != nil {
		return nil, "", err
	}

	chunk, err := io.ReadAll(file)

	if err != nil {
		return nil, "", err
	}
	objectInfo, err := t.minio.StatObject(ctx, userTweetsName, fileName, minio.StatObjectOptions{})

	if err != nil {
		return nil, "", err
	}

	return chunk, objectInfo.ContentType, nil
}

func (t *TweetMinio) UpdateTweetImage(ctx context.Context, oldName string, newName string) error {
	copyDestOpts := minio.CopyDestOptions{
		Bucket: userTweetsName,
		Object: newName,
	}

	copySrcOpts := minio.CopySrcOptions{
		Bucket: userTweetsName,
		Object: oldName,
	}

	if _, err := t.minio.CopyObject(ctx, copyDestOpts, copySrcOpts); err != nil {
		return err
	}

	if err := t.minio.RemoveObject(ctx, userTweetsName, oldName, minio.RemoveObjectOptions{}); err != nil {
		return err
	}

	return nil
}

func (t *TweetMinio) DeleteFile(ctx context.Context, fileName string) error {
	if err := t.minio.RemoveObject(ctx, userTweetsName, fileName, minio.RemoveObjectOptions{}); err != nil {
		return err
	}

	return nil
}
