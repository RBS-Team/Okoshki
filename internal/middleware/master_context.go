package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	usersDTO "github.com/RBS-Team/Okoshki/microservices/core/users/dto"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

// MasterLookup — минимальный контракт для резолва мастера по user_id.
// Реализуется users.service.Service.GetMasterByUserID.
type MasterLookup interface {
	GetMasterByUserID(ctx context.Context, userID uuid.UUID) (*usersDTO.Master, error)
}

const masterIDKey ctxKey = "master_id"

// MasterContext — middleware, который по user_id из JWT-claims резолвит master_id и кладёт в контекст.
// Применяется к эндпоинтам, где залогиненный мастер действует над «своими» ресурсами (/me/...).
//
// Предусловие: AuthMiddleware уже отработал и положил claims.
// Если пользователь не является мастером — 403.
func MasterContext(lookup MasterLookup) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "middleware.MasterContext"
			log := LoggerFromContext(r.Context())

			userIDStr, ok := GetUserID(r.Context())
			if !ok || userIDStr == "" {
				log.Warnf("[%s]: missing user id in context", op)
				response.UnauthorizedJSON(w)
				return
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				log.Warnf("[%s]: invalid user id format: %v", op, err)
				response.UnauthorizedJSON(w)
				return
			}

			master, err := lookup.GetMasterByUserID(r.Context(), userID)
			if err != nil {
				log.Warnf("[%s]: cannot resolve master for user %s: %v", op, userIDStr, err)
				response.ForbiddenJSON(w)
				return
			}

			masterID, err := uuid.Parse(master.ID)
			if err != nil {
				log.Errorf("[%s]: invalid master id from lookup: %v", op, err)
				response.InternalErrorJSON(w)
				return
			}

			ctx := context.WithValue(r.Context(), masterIDKey, masterID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// MasterIDFromContext возвращает master_id, положенный MasterContext.
func MasterIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(masterIDKey).(uuid.UUID)
	return v, ok
}
