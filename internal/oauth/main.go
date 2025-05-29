package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"scoreboard-api/internal/user"
)

type Handler struct {
	logger      *zap.Logger
	userService *user.Service
	oauthConfig *oauth2.Config
}

func NewHandler(logger *zap.Logger, userService *user.Service, oauthConfig *oauth2.Config) *Handler {
	return &Handler{
		logger:      logger,
		userService: userService,
		oauthConfig: oauthConfig,
	}
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	logger.Info("hello world", zap.String("hello", "world"))

	dbpool, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5432/postgres")
	if err != nil {
		panic(err)
	}
	defer dbpool.Close()

	userService := user.NewService(logger, dbpool)
	oauthConfig := &oauth2.Config{
		ClientID:     "883905598480-anv2pnkpl684u7g1hv5rh4b508qqfhb8.apps.googleusercontent.com",
		ClientSecret: "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",
		RedirectURL:  "http://localhost:8080/api/oauth/google/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	handler := NewHandler(logger, userService, oauthConfig)
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handler.healthz)

	mux.HandleFunc("GET /api/login/oauth/google", handler.oauth2Start)
	mux.HandleFunc("GET /api/oauth/google/callback", handler.oauth2Callback)
	mux.HandleFunc("GET /frontend", handler.frontend)

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		logger.Error("Failed to start server", zap.Error(err))
		return
	}
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")
}

func (h *Handler) oauth2Start(w http.ResponseWriter, r *http.Request) {
	// Placeholder for OAuth2 start logic
	callback := r.URL.Query().Get("c")
	redirectTo := r.URL.Query().Get("r")
	if callback == "" {
		callback = fmt.Sprintf("%s/api/oauth/debug/token", "http://localhost:8080")
	}
	if redirectTo != "" {
		callback = fmt.Sprintf("%s?r=%s", callback, redirectTo)
	}

	state := base64.StdEncoding.EncodeToString([]byte(callback))

	authURL := h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *Handler) oauth2Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	oauthError := r.URL.Query().Get("error") // Check if there was an error during the OAuth2 process

	callbackURL, err := base64.StdEncoding.DecodeString(state)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?error=%s", string(callbackURL), "Invalid state"), http.StatusTemporaryRedirect)
		return
	}

	callback, err := url.Parse(string(callbackURL))
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?error=%s", string(callbackURL), "Invalid callback URL"), http.StatusTemporaryRedirect)
		return
	}

	redirectTo := callback.Query().Get("r")
	callback.RawQuery = ""

	if oauthError != "" {
		http.Redirect(w, r, fmt.Sprintf("%s?error=%s", callback, oauthError), http.StatusTemporaryRedirect)
		return
	}

	token, err := h.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?error=%s", callback, err.Error()), http.StatusTemporaryRedirect)
	}

	userInfo, err := getUserInfo(h.oauthConfig, token)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?error=%s", callback, err.Error()), http.StatusTemporaryRedirect)
	}

	dbUser, err := h.userService.FindOrCreateWithProfile(r.Context(), 
		userInfo.Email,
		userInfo.Name,
		userInfo.GivenName,
		userInfo.FamilyName,
		userInfo.Picture,
		userInfo.Locale,
		userInfo.EmailVerified,
	)
	if err != nil {
		h.logger.Error("Failed to find or create user", zap.Error(err))
		http.Redirect(w, r, fmt.Sprintf("%s?error=%s", callback, err.Error()), http.StatusTemporaryRedirect)
		return
	}

	userInfoJSON, err := json.Marshal(map[string]interface{}{
		"user":   userInfo,
		"dbUser": dbUser,
	})
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?error=%s", callback, err.Error()), http.StatusTemporaryRedirect)
	}

	base64Json := base64.StdEncoding.EncodeToString(userInfoJSON)
	escaped := url.QueryEscape(base64Json)

	var redirectWithToken string
	if redirectTo != "" {
		redirectWithToken = fmt.Sprintf("%s?token=%s&r=%s", callback, escaped, redirectTo)
	} else {
		redirectWithToken = fmt.Sprintf("%s?token=%s", callback, escaped)
	}

	http.Redirect(w, r, redirectWithToken, http.StatusTemporaryRedirect)
}

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Locale        string `json:"locale"`
}

func getUserInfo(config *oauth2.Config, token *oauth2.Token) (googleUserInfo, error) {
	client := config.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return googleUserInfo{}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	var userInfo googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return googleUserInfo{}, err
	}

	return userInfo, nil
}

func (h *Handler) frontend(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	page := `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>OAuth Callback</title>
  <style>
    body { font-family: sans-serif; padding: 2em; }
    pre { background: #f0f0f0; padding: 1em; border-radius: 5px; white-space: pre-wrap; word-break: break-all; }
  </style>
</head>
<body>
  <h2>OAuth Login Result</h2>
  <div id="output">Loading...</div>

  <script>
    function getQueryParam(name) {
      const params = new URLSearchParams(window.location.search);
      return params.get(name);
    }

    function base64DecodeUnicode(str) {
      try {
        // 把 base64 字串 decode 成 UTF-8
        const decoded = atob(str.replace(/ /g, '+'));
        return decodeURIComponent(escape(decoded));
      } catch (e) {
        return 'Failed to decode token: ' + e.message;
      }
    }

    function prettyPrintJSON(str) {
      try {
        const obj = JSON.parse(str);
        return JSON.stringify(obj, null, 2);
      } catch (e) {
        return 'Invalid JSON: ' + e.message + '\\n\\n' + str;
      }
    }

    const token = getQueryParam('token');
    const redirectTo = getQueryParam('r');

    let html = '';

    if (!token) {
      html += '<p>Not logged in.</p>';
      html += '<p><a href="/api/login/oauth/google?c=http://localhost:8080/frontend">Login with Google</a></p>';
    } else {
      const jsonText = base64DecodeUnicode(token);
      html += '<p><strong>User Info:</strong></p>';
      html += '<pre>' + prettyPrintJSON(jsonText) + '</pre>';

      if (redirectTo) {
        html += '<p><strong>Redirect Target:</strong> ' + redirectTo + '</p>';
      }
    }

    document.getElementById('output').innerHTML = html;
  </script>
</body>
</html>
`

	_, _ = w.Write([]byte(page))
}
