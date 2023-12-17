package domain

import (
	"github.com/google/uuid"
	"time"
)

type Tweet struct {
	TweetID   uuid.UUID `json:"tweet_id" db:"tweet_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Text      string    `json:"text" db:"text"`
	ImageName string    `json:"image" db:"image_name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
