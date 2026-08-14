package database

import (
	"context"
	"database/sql"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"time"
)

// type MetadataEvent struct {
// 	ID         int64  `json:"id"`
// 	Event      string `json:"event"`
// 	Repository string `json:"repository"`
// 	ReceivedAt string `json:"received_at"`
// }
type MetadataEvent struct {
    ID         int64  `json:"id"`
    EventType  string `json:"event_type"`
    Project    string `json:"project"`
    Repository string `json:"repository"`
    Tag        string `json:"tag"`
    Digest     string `json:"digest"`
    Operator   string `json:"operator"`
    OccurredAt time.Time `json:"occurred_at"`
    ReceivedAt time.Time `json:"received_at"`
}

func GetEvents(
	ctx context.Context,
	db *sql.DB,
) ([]MetadataEvent, error) {

	tracer := otel.Tracer("metadata-api")

ctx, span := tracer.Start(ctx, "database.query_events")
defer span.End()

// rows, err := db.QueryContext(ctx, `
// 	SELECT id, event, repository, received_at
// 	FROM metadata_events
// 	ORDER BY id;
// `)
rows, err := db.QueryContext(ctx, `
    SELECT
        id,
        event_type,
        project,
        repository,
        tag,
        digest,
        operator,
        occurred_at,
        received_at
    FROM metadata_events
    ORDER BY id;
`)
	if err != nil {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return nil, err
}
	defer rows.Close()

	var events []MetadataEvent

	// for rows.Next() {
	// 	var event MetadataEvent

	// 	if err := rows.Scan(
	// 		&event.ID,
	// 		&event.Event,
	// 		&event.Repository,
	// 		&event.ReceivedAt,
	// 	); err != nil {
	// 		return nil, err
	// 	}

	// 	events = append(events, event)
	// }

	for rows.Next() {
    var event MetadataEvent

    if err := rows.Scan(
        &event.ID,
        &event.EventType,
        &event.Project,
        &event.Repository,
        &event.Tag,
        &event.Digest,
        &event.Operator,
        &event.OccurredAt,
        &event.ReceivedAt,
    ); err != nil {
        return nil, err
    }

    events = append(events, event)
}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}