package scoreboard

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Service struct {
	logger *zap.Logger
	tracer trace.Tracer
	query  Querier
}

type Querier interface {
	List(ctx context.Context) ([]Scoreboard, error)
	Get(ctx context.Context, id uuid.UUID) (Scoreboard, error)
	Create(ctx context.Context, name pgtype.Text) (Scoreboard, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, arg UpdateParams) (Scoreboard, error)
}

func NewService(logger *zap.Logger, db *pgxpool.Pool) *Service {
	return &Service{
		logger: logger,
		tracer: otel.Tracer("scoreboard/service"),
		query:  New(db),
	}
}

func (s Service) List(ctx context.Context) ([]Scoreboard, error) {
	traceCtx, span := s.tracer.Start(ctx, "GetAll")
	defer span.End()
	scoreboard, err := s.query.List(traceCtx)
	if err != nil {
		return nil, err
	}
	return scoreboard, nil
}

func (s Service) Get(ctx context.Context, id uuid.UUID) (Scoreboard, error) {
	traceCtx, span := s.tracer.Start(ctx, "GetByID")
	defer span.End()
	scoreboard, err := s.query.Get(traceCtx, id)
	if err != nil {
		return Scoreboard{}, err
	}
	return scoreboard, nil
}

func (s Service) Create(ctx context.Context, name pgtype.Text) (Scoreboard, error) {
	traceCtx, span := s.tracer.Start(ctx, "Create")
	defer span.End()
	createdScoreboard, err := s.query.Create(traceCtx, name)
	if err != nil {
		return Scoreboard{}, err
	}
	return createdScoreboard, nil
}

func (s Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.query.Delete(ctx, id)
}

func (s Service) Update(ctx context.Context, arg UpdateParams) (Scoreboard, error) {
	return s.query.Update(ctx, arg)
}
