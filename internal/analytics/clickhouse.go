package analytics

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ClickHouseWriter struct {
	conn driver.Conn
}

func NewClickHouseWriter(addr, database, username, password string) (*ClickHouseWriter, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
	})
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &ClickHouseWriter{conn: conn}, nil
}

func (w *ClickHouseWriter) Track(ctx context.Context, event Event) error {
	return w.conn.Exec(ctx, `
		INSERT INTO analytics_events
			(event_time, event_type, entity_type, entity_id, department_id, metadata)
		VALUES
			(?, ?, ?, ?, ?, ?)
	`,
		event.Time,
		event.Type,
		event.EntityType,
		event.EntityID,
		event.DepartmentID,
		event.Metadata,
	)
}
