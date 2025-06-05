package scoreboard

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Store interface {
	List(ctx context.Context) ([]Scoreboard, error)
	Get(ctx context.Context, id uuid.UUID) (Scoreboard, error)
	Create(ctx context.Context, name pgtype.Text) (Scoreboard, error)
	Update(ctx context.Context, arg UpdateParams) (Scoreboard, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// CreateScoreboardPayload defines the expected request body for creating a scoreboard.
type CreateScoreboardPayload struct {
	Name string `json:"name" validate:"required,Alphanumericspaceunderhyphen"`
}

type Response struct {
	ID        string `json:"id" validate:"required,uuid4"`
	Name      string `json:"name" validate:"required"`
	CreatedAt string `json:"createdAt" validate:"required"`
	UpdatedAt string `json:"updatedAt" validate:"required"`
}

type Handler struct {
	validator *validator.Validate
	tracer    trace.Tracer
	logger    *zap.Logger
	store     Store
}

func NewHandler(v *validator.Validate, logger *zap.Logger, s Store) Handler {
	return Handler{
		validator: v,
		tracer:    otel.Tracer("Scoreboard/handler"),
		logger:    logger,
		store:     s,
	}
}

func (h Handler) ListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	scoreboards, err := h.store.List(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := make([]Response, len(scoreboards))
	for index, post := range scoreboards {
		response[index] = GenerateResponse(post)
	}
	WriteJSONResponse(w, http.StatusOK, response)
}

func (h Handler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var payload CreateScoreboardPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)
	name := pgtype.Text{
		String: payload.Name,
		Valid:  payload.Name != "",
	}
	scoreboard, err := h.store.Create(ctx, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := GenerateResponse(scoreboard)
	WriteJSONResponse(w, http.StatusOK, response)
}

func (h Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Extract ID from the URL path
	path := r.URL.Path
	// Remove the trailing slash if present
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	// Get the last segment of the path
	segments := strings.Split(path, "/")
	if len(segments) < 1 {
		http.Error(w, "Invalid path for ID extraction", http.StatusBadRequest)
		return
	}
	idStr := segments[len(segments)-1]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid UUID format", http.StatusBadRequest)
		return
	}
	scoreboard, err := h.store.Get(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := GenerateResponse(scoreboard)
	WriteJSONResponse(w, http.StatusOK, response)
}

func (h Handler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract ID from the URL path
	path := r.URL.Path
	// Remove the trailing slash if present
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	// Get the last segment of the path
	segments := strings.Split(path, "/")
	if len(segments) < 1 {
		http.Error(w, "Invalid path for ID extraction", http.StatusBadRequest)
		return
	}
	idStr := segments[len(segments)-1]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid UUID format", http.StatusBadRequest)
		return
	}

	var payload CreateScoreboardPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	name := pgtype.Text{
		String: payload.Name,
		Valid:  payload.Name != "",
	}
	scoreboard, err := h.store.Update(ctx, UpdateParams{
		ID:   id,
		Name: name,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := GenerateResponse(scoreboard)
	WriteJSONResponse(w, http.StatusOK, response)
}

func (h Handler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract ID from the URL path
	path := r.URL.Path
	// Remove the trailing slash if present
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	// Get the last segment of the path
	segments := strings.Split(path, "/")
	if len(segments) < 1 {
		http.Error(w, "Invalid path for ID extraction", http.StatusBadRequest)
		return
	}
	idStr := segments[len(segments)-1]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid UUID format", http.StatusBadRequest)
		return
	}

	err = h.store.Delete(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func WriteJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	_, err = w.Write(jsonBytes)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func GenerateResponse(scoreboard Scoreboard) Response {
	return Response{
		ID:        scoreboard.ID.String(),
		Name:      scoreboard.Name.String,
		CreatedAt: scoreboard.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: scoreboard.UpdatedAt.Time.Format(time.RFC3339),
	}
}
