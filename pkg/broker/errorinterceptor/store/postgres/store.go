package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/broker/errorinterceptor/entity"
	"kafka-polygon/pkg/cerror"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

type Store struct {
	db *bun.DB
}

func NewStore(db *bun.DB) *Store {
	return &Store{db: db}
}

type failedBrokerEvent struct {
	bun.BaseModel `bun:"table:failed_broker_event"`
	ID            string          `bun:"id,pk"`
	CreatedAt     time.Time       `bun:"created_at"`
	UpdatedAt     time.Time       `bun:"updated_at"`
	Status        string          `bun:"status"`
	Topic         string          `bun:"topic"`
	Data          json.RawMessage `bun:"data,type:jsonb"`
	Error         string          `bun:"error"`
}

func (r *Store) SaveEvent(ctx context.Context, e *entity.FailedBrokerEvent) (*entity.FailedBrokerEvent, error) {
	dst := &failedBrokerEvent{
		ID:        e.ID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Status:    e.Status,
		Data:      e.Data,
		Topic:     e.Topic,
		Error:     e.Error,
	}

	_, err := r.db.NewInsert().
		Model(dst).
		On("CONFLICT (id) DO UPDATE").
		Set(onConflictUpdateValuesFromColumns([]string{"updated_at", "status", "topic", "data", "error"})).
		Exec(ctx)
	if err != nil {
		return nil, cerror.NewF(ctx,
			cerror.DBToKind(err),
			"create failed broker event --> %+v", err).LogError()
	}

	return &entity.FailedBrokerEvent{
		ID:        dst.ID,
		CreatedAt: dst.CreatedAt,
		UpdatedAt: dst.UpdatedAt,
		Status:    dst.Status,
		Topic:     dst.Topic,
		Data:      dst.Data,
		Error:     dst.Error,
	}, nil
}

func (r *Store) SearchEvents(ctx context.Context) ([]*entity.FailedBrokerEvent, error) {
	events := make([]*failedBrokerEvent, 0)
	err := r.db.NewSelect().Model(&events).Order("created_at").Scan(ctx)

	if err != nil {
		return nil, cerror.NewF(ctx,
			cerror.DBToKind(err),
			"get all failed broker events --> %+v", err).LogError()
	}

	result := make([]*entity.FailedBrokerEvent, len(events))

	for i := range events {
		result[i] = &entity.FailedBrokerEvent{
			ID:        events[i].ID,
			CreatedAt: events[i].CreatedAt,
			Data:      events[i].Data,
			Error:     events[i].Error,
		}
	}

	return result, nil
}

func onConflictUpdateValuesFromColumns(columns []string) string {
	parts := make([]string, len(columns))

	for i, c := range columns {
		parts[i] = fmt.Sprintf("%s=EXCLUDED.%s", c, c)
	}

	return strings.Join(parts, ",")
}
