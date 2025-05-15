package scoreboard

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

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

type Response struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
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

func validateName(name string) (bool, string) {
	if name == "" {
		return false, "Name cannot be empty"
	}
	if len(name) > 255 {
		return false, "Name cannot exceed 255 characters"
	}
	validNamePattern := regexp.MustCompile(`^[a-zA-Z0-9\-_ ]+$`)
	if !validNamePattern.MatchString(name) {
		return false, "Name can only contain alphanumeric characters, hyphens, underscores, and spaces"
	}

	return true, ""
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
	defer r.Body.Close()

	if valid, errMsg := validateName(payload.Name); !valid {
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	scoreboard, err := h.store.Create(ctx, payload.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := GenerateResponse(scoreboard)
	WriteJSONResponse(w, http.StatusOK, response)
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
	response := GenerateResponse(scoreboard)
	WriteJSONResponse(w, http.StatusOK, response)
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

	if valid, errMsg := validateName(payload.Name); !valid {
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	scoreboard, err := h.store.Update(ctx, UpdateParams{
		ID:   id,
		Name: payload.Name,
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
		Name:      scoreboard.Name,
		CreatedAt: scoreboard.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: scoreboard.UpdatedAt.Time.Format(time.RFC3339),
	}
}
