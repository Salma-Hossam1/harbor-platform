package database

import "database/sql"

type HarborEvent struct {
	Type       string          `json:"type"`
	OccurredAt int64           `json:"occur_at"`
	Operator   string          `json:"operator"`
	EventData  HarborEventData `json:"event_data"`
}

type HarborEventData struct {
	Resources  []HarborResource `json:"resources"`
	Repository HarborRepository `json:"repository"`
}

type HarborResource struct {
	Digest      string `json:"digest"`
	Tag         string `json:"tag"`
	ResourceURL string `json:"resource_url"`
}

type HarborRepository struct {
	DateCreated  int64  `json:"date_created"`
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	RepoFullName string `json:"repo_full_name"`
	RepoType     string `json:"repo_type"`
}

type MetadataEvent struct {
	EventType  string
	Project    string
	Repository string
	Tag        string
	Digest     string
	Operator   string
	OccurredAt int64
}

const insertEvent = `
INSERT INTO metadata_events (
    event_type,
    project,
    repository,
    tag,
    digest,
    operator,
    occurred_at
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    to_timestamp($7)
);
`

func InsertEvent(db *sql.DB, event MetadataEvent) error {
	_, err := db.Exec(
		insertEvent,
		event.EventType,
		event.Project,
		event.Repository,
		event.Tag,
		event.Digest,
		event.Operator,
		event.OccurredAt,
	)

	return err
}
