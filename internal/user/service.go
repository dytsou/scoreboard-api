package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Service struct {
	queries *Queries
	logger  *zap.Logger
	tracer  trace.Tracer
}

func NewService(logger *zap.Logger, db DBTX) *Service {
	return &Service{
		queries: New(db),
		logger:  logger,
		tracer:  otel.Tracer("user/service"),
	}
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	traceCtx, span := s.tracer.Start(ctx, "GetByID")
	defer span.End()
	user, err := s.queries.GetByID(traceCtx, id)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *Service) GetByEmail(ctx context.Context, email string) (User, error) {
	traceCtx, span := s.tracer.Start(ctx, "GetByEmail")
	defer span.End()
	user, err := s.queries.GetByEmail(traceCtx, email)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

// Create user with basic email only (for backward compatibility)
func (s *Service) Create(ctx context.Context, email string) (User, error) {
	traceCtx, span := s.tracer.Start(ctx, "Create")
	defer span.End()
	
	params := CreateParams{
		Email:         email,
		Name:          pgtype.Text{String: "", Valid: false},
		GivenName:     pgtype.Text{String: "", Valid: false},
		FamilyName:    pgtype.Text{String: "", Valid: false},
		Picture:       pgtype.Text{String: "", Valid: false},
		EmailVerified: false,
		Locale:        pgtype.Text{String: "", Valid: false},
	}
	
	user, err := s.queries.Create(traceCtx, params)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *Service) ExistByEmail(ctx context.Context, email string) (bool, error) {
	traceCtx, span := s.tracer.Start(ctx, "ExistByEmail")
	defer span.End()
	user, err := s.queries.ExistsByEmail(traceCtx, email)
	if err != nil {
		return false, err
	}
	return user, nil
}

func (s *Service) CreateWithProfile(ctx context.Context, email, name, givenName, familyName, picture, locale string, emailVerified bool) (User, error) {
	traceCtx, span := s.tracer.Start(ctx, "CreateWithProfile")
	defer span.End()
	
	params := CreateParams{
		Email:         email,
		Name:          pgtype.Text{String: name, Valid: name != ""},
		GivenName:     pgtype.Text{String: givenName, Valid: givenName != ""},
		FamilyName:    pgtype.Text{String: familyName, Valid: familyName != ""},
		Picture:       pgtype.Text{String: picture, Valid: picture != ""},
		EmailVerified: emailVerified,
		Locale:        pgtype.Text{String: locale, Valid: locale != ""},
	}
	
	user, err := s.queries.Create(traceCtx, params)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *Service) UpdateProfile(ctx context.Context, email, name, givenName, familyName, picture, locale string, emailVerified bool) (User, error) {
	traceCtx, span := s.tracer.Start(ctx, "UpdateProfile")
	defer span.End()
	
	params := UpdateParams{
		Email:         email,
		Name:          pgtype.Text{String: name, Valid: name != ""},
		GivenName:     pgtype.Text{String: givenName, Valid: givenName != ""},
		FamilyName:    pgtype.Text{String: familyName, Valid: familyName != ""},
		Picture:       pgtype.Text{String: picture, Valid: picture != ""},
		EmailVerified: emailVerified,
		Locale:        pgtype.Text{String: locale, Valid: locale != ""},
	}
	
	user, err := s.queries.Update(traceCtx, params)
	if err != nil {
		return User{}, err
	}
	return user, nil
}



func (s *Service) FindOrCreateWithProfile(ctx context.Context, email, name, givenName, familyName, picture, locale string, emailVerified bool) (User, error) {
	traceCtx, span := s.tracer.Start(ctx, "FindOrCreateWithProfile")
	defer span.End()
	
	exist, err := s.ExistByEmail(traceCtx, email)
	if err != nil {
		return User{}, err
	}
	
	if exist {
		user, err := s.GetByEmail(traceCtx, email)
		if err != nil {
			return User{}, err
		}
		
		// Update the user's profile with latest information from OAuth provider
		updatedUser, err := s.UpdateProfile(traceCtx, email, name, givenName, familyName, picture, locale, emailVerified)
		if err != nil {
			// If update fails, return the existing user
			s.logger.Warn("Failed to update user profile", zap.Error(err), zap.String("email", email))
			return user, nil
		}
		return updatedUser, nil
	}
	
	user, err := s.CreateWithProfile(traceCtx, email, name, givenName, familyName, picture, locale, emailVerified)
	if err != nil {
		return User{}, err
	}
	return user, nil
}