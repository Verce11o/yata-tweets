package postgres

import (
	"context"
	"database/sql"
	pb "github.com/Verce11o/yata-protos/gen/go/tweets"
	"github.com/Verce11o/yata-tweets/internal/domain"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"
)

type TweetPostgres struct {
	db     *sqlx.DB
	tracer trace.Tracer
}

func NewTweetPostgres(db *sqlx.DB, tracer trace.Tracer) *TweetPostgres {
	return &TweetPostgres{db: db, tracer: tracer}
}

func (t *TweetPostgres) CreateTweet(ctx context.Context, input *pb.CreateTweetRequest, imageName string) (string, error) {
	ctx, span := t.tracer.Start(ctx, "tweetPostgres.CreateTweet")
	defer span.End()

	var tweetID string

	q := "INSERT INTO tweets (user_id, text, image) VALUES ($1, $2, $3) RETURNING tweet_id"

	stmt, err := t.db.PreparexContext(ctx, q)

	if err != nil {
		return "", err
	}

	err = stmt.QueryRowxContext(ctx, input.GetUserId(), input.GetText(), imageName).Scan(&tweetID)

	if err != nil {
		return "", err
	}

	return tweetID, nil

}

func (t *TweetPostgres) GetTweet(ctx context.Context, tweetID string) (*domain.Tweet, error) {
	ctx, span := t.tracer.Start(ctx, "tweetPostgres.GetTweet")
	defer span.End()

	var tweet domain.Tweet

	q := "SELECT * FROM tweets WHERE tweet_id = $1"

	err := t.db.QueryRowxContext(ctx, q, tweetID).StructScan(&tweet)

	if err != nil {
		return nil, sql.ErrNoRows
	}

	return &tweet, nil
}

func (t *TweetPostgres) UpdateTweet(ctx context.Context, input *pb.UpdateTweetRequest, imageName string) (*domain.Tweet, error) {
	ctx, span := t.tracer.Start(ctx, "tweetPostgres.UpdateTweet")
	defer span.End()

	var tweet domain.Tweet

	q := "UPDATE tweets SET text = $1, image = $2, updated_at = CURRENT_TIMESTAMP WHERE tweet_id = $3 RETURNING *"

	if err := t.db.QueryRowxContext(ctx, q, input.GetText(), imageName, input.GetTweetId()).StructScan(&tweet); err != nil {
		return nil, err
	}

	return &tweet, nil

}

func (t *TweetPostgres) DeleteTweet(ctx context.Context, tweetID string) error {
	ctx, span := t.tracer.Start(ctx, "tweetPostgres.DeleteTweet")
	defer span.End()

	q := "DELETE FROM tweets WHERE tweet_id = $1"

	res, err := t.db.ExecContext(ctx, q, tweetID)

	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
