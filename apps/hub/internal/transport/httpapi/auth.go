package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	api "github.com/harmost/api/gen"
	"github.com/harmost/hub/internal/auth"
	"github.com/harmost/hub/internal/domain"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

// ─── GitHub OAuth ─────────────────────────────────────────────────────────────

func (s *Server) githubOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.cfg.GitHubClientID,
		ClientSecret: s.cfg.GitHubClientSecret,
		RedirectURL:  s.cfg.GitHubCallbackURL,
		Scopes:       []string{"read:user", "user:email"},
		Endpoint:     githuboauth.Endpoint,
	}
}

func (s *Server) handleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	state, err := randomHex(8)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "internal error")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
		Path:     "/",
	})
	http.Redirect(w, r, s.githubOAuthConfig().AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func (s *Server) handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state.
	cookie, err := r.Cookie("oauth_state")
	if err != nil || cookie.Value != r.URL.Query().Get("state") {
		jsonError(w, http.StatusBadRequest, "invalid oauth state")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", MaxAge: -1, Path: "/"})

	// Exchange code for token.
	code := r.URL.Query().Get("code")
	ghToken, err := s.githubOAuthConfig().Exchange(r.Context(), code)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "failed to exchange oauth code")
		return
	}

	// Fetch GitHub profile.
	profile, err := fetchGitHubProfile(r.Context(), ghToken.AccessToken)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to fetch github profile")
		return
	}

	// Upsert user + org.
	user, org, err := s.svc.User.SignUpOrLogin(r.Context(), *profile)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to sign in")
		return
	}

	// Issue JWT.
	token, err := auth.Sign(user.ID, org.ID, s.cfg.JWTSecret)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to issue token")
		return
	}

	redirectURL := fmt.Sprintf("%s/auth/callback?token=%s", s.cfg.FrontendURL, url.QueryEscape(token))
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

type githubAPIUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

func fetchGitHubProfile(ctx context.Context, accessToken string) (*domain.GitHubProfile, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var gu githubAPIUser
	if err := json.Unmarshal(body, &gu); err != nil {
		return nil, err
	}

	email := gu.Email
	if email == "" {
		email = fmt.Sprintf("%d+%s@users.noreply.github.com", gu.ID, gu.Login)
	}
	name := gu.Name
	if name == "" {
		name = gu.Login
	}

	return &domain.GitHubProfile{
		GitHubID:  fmt.Sprintf("%d", gu.ID),
		Email:     email,
		Name:      name,
		AvatarURL: gu.AvatarURL,
		Login:     gu.Login,
	}, nil
}

// ─── Device flow ──────────────────────────────────────────────────────────────

func (s *Server) handleDeviceAuthorize(w http.ResponseWriter, r *http.Request) {
	d, err := s.svc.DeviceFlow.Initiate(r.Context())
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to initiate device flow")
		return
	}
	jsonOK(w, api.DeviceAuthorizeResponse{
		DeviceCode:      d.DeviceCode,
		UserCode:        d.UserCode,
		VerificationURI: fmt.Sprintf("%s/device?code=%s", s.cfg.FrontendURL, url.QueryEscape(d.UserCode)),
		ExpiresIn:       int(time.Until(d.ExpiresAt).Seconds()),
		Interval:        5,
		GRPCAddr:        s.cfg.GRPCAddr,
	})
}

func (s *Server) handleDeviceToken(w http.ResponseWriter, r *http.Request) {
	var req api.DeviceTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DeviceCode == "" {
		jsonError(w, http.StatusBadRequest, "missing device_code")
		return
	}

	token, pending, err := s.svc.DeviceFlow.Poll(r.Context(), req.DeviceCode)
	if err != nil {
		jsonError(w, http.StatusGone, "expired or invalid device code")
		return
	}
	if pending {
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(api.Error{Error: "authorization_pending"})
		return
	}
	jsonOK(w, api.DeviceTokenResponse{
		AccessToken: token,
		TokenType:   api.TokenTypeBearer,
	})
}

func (s *Server) handleDeviceApprove(w http.ResponseWriter, r *http.Request) {
	claims := claimsFromCtx(r.Context())

	var req api.DeviceApproveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserCode == "" {
		jsonError(w, http.StatusBadRequest, "missing user_code")
		return
	}

	if err := s.svc.DeviceFlow.Approve(r.Context(), req.UserCode, claims.OrgID, claims.UserID); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	var sb strings.Builder
	for _, c := range b {
		fmt.Fprintf(&sb, "%02x", c)
	}
	return sb.String(), nil
}
