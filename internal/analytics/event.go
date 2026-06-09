package analytics

import (
	"context"
	"time"
)

type Event struct {
	Time         time.Time
	Type         string
	EntityType   string
	EntityID     uint64
	DepartmentID *uint64
	Metadata     string
}

type Writer interface {
	Track(ctx context.Context, event *Event) error
}
