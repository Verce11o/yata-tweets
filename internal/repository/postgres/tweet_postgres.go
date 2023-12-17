package postgres

import (
	"context"
	"database/sql"
	pb "github.com/Verce11o/yata-protos/gen/go/tweets"
	"github.com/Verce11o/yata-tweets/internal/domain"
	"github.com/Verce11o/yata-tweets/internal/lib/pagination"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"
	"time"
)

const (
	paginationLimit = 10
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

	q := "INSERT INTO tweets (user_id, text, image_name) VALUES ($1, $2, $3) RETURNING tweet_id"

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

func (t *TweetPostgres) GetAllTweets(ctx context.Context, cursor string) ([]*pb.Tweet, string, error) {
	ctx, span := t.tracer.Start(ctx, "tweetPostgres.GetAllTweets")
	defer span.End()

	var createdAt time.Time
	var tweetID uuid.UUID
	var err error

	if cursor != "" {
		createdAt, tweetID, err = pagination.DecodeCursor(cursor)
		if err != nil {
			return nil, "", err
		}
	}

	q := "SELECT * FROM tweets WHERE (created_at, tweet_id) > ($1, $2) ORDER BY created_at, tweet_id LIMIT $3"

	rows, err := t.db.QueryxContext(ctx, q, createdAt, tweetID, paginationLimit)

	if err != nil {
		return nil, "", err
	}

	var tweets []*pb.Tweet
	var latestCreatedAt time.Time

	for rows.Next() {
		var item domain.Tweet
		err = rows.StructScan(&item)
		if err != nil {
			return nil, "", err
		}
		tweets = append(tweets, &pb.Tweet{
			UserId:  item.UserID.String(),
			TweetId: item.TweetID.String(),
			Text:    item.Text,
		})
		latestCreatedAt = item.CreatedAt
	}

	var nextCursor string
	if len(tweets) > 0 {
		nextCursor = pagination.EncodeCursor(latestCreatedAt, tweets[len(tweets)-1].TweetId)
	}

	return tweets, nextCursor, nil
}

func (t *TweetPostgres) UpdateTweet(ctx context.Context, input *pb.UpdateTweetRequest, imageName string) (*domain.Tweet, error) {
	ctx, span := t.tracer.Start(ctx, "tweetPostgres.UpdateTweet")
	defer span.End()

	var tweet domain.Tweet

	q := "UPDATE tweets SET text = $1, image_name = $2, updated_at = CURRENT_TIMESTAMP WHERE tweet_id = $3 RETURNING *"

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
