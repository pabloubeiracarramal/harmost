package httpapi

import (
	"errors"
	"net/http"

	api "github.com/harmost/api/gen"
	"github.com/harmost/hub/internal/domain"
)

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
	jsonOK(w, api.User{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		OrgID:     claims.OrgID,
	})
}
