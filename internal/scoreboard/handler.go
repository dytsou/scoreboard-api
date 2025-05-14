package scoreboard

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Store interface {
	List(ctx context.Context) ([]Scoreboard, error)
	Get(ctx context.Context, id uuid.UUID) (Scoreboard, error)
	Create(ctx context.Context, name string) (Scoreboard, error)
	Update(ctx context.Context, arg UpdateParams) (Scoreboard, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// CreateScoreboardPayload defines the expected request body for creating a scoreboard.
type CreateScoreboardPayload struct {
	Name string `json:"name"`
}

type Handler struct {
	tracer trace.Tracer
	logger *zap.Logger
	store  Store
}

func NewHandler(logger *zap.Logger, s Store) Handler {
	return Handler{
		tracer: otel.Tracer("Scoreboard/handler"),
		logger: logger,
		store:  s,
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
	json.NewEncoder(w).Encode(scoreboards)
}

func (h Handler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var payload CreateScoreboardPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if payload.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	scoreboard, err := h.store.Create(ctx, payload.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(scoreboard)
}

func (h Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Extract ID from URL path
	path := r.URL.Path
	// Remove trailing slash if present
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(scoreboard)
}

func (h Handler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract ID from URL path
	path := r.URL.Path
	// Remove trailing slash if present
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
	defer r.Body.Close()

	scoreboard, err := h.store.Update(ctx, UpdateParams{
		ID:   id,
		Name: payload.Name,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(scoreboard)
}

func (h Handler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract ID from URL path
	path := r.URL.Path
	// Remove trailing slash if present
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
