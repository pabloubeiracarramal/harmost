package httpapi

import (
	"errors"
	"net/http"

	"github.com/harmost/hub/internal/domain"
)

type userResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	OrgID     string `json:"org_id"`
}

func (s *Server) getMe(w http.ResponseWriter, r *http.Request) {
	claims := claimsFromCtx(r.Context())
	user, err := s.svc.User.GetByID(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, http.StatusUnauthorized, "user not found")
			return
		}
		jsonError(w, http.StatusInternalServerError, "failed to get user")
		return
	}
	jsonOK(w, userResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		OrgID:     claims.OrgID,
	})
}
