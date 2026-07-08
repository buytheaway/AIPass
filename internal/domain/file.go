package domain

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID           uuid.UUID `db:"id" json:"id"`
	Bucket       string    `db:"bucket" json:"bucket"`
	ObjectKey    string    `db:"object_key" json:"object_key"`
	OriginalName string    `db:"original_name" json:"original_name"`
	ContentType  string    `db:"content_type" json:"content_type"`
	SizeBytes    int64     `db:"size_bytes" json:"size_bytes"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
