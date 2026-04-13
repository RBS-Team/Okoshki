package middleware

import (
	"net/http"

	"github.com/RBS-Team/Okoshki/pkg/response"
)

// RequireRole проверяет, есть ли у пользователя нужная роль
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "middleware.RequireRole"
			log := LoggerFromContext(r.Context())

			userRole, ok := GetUserRole(r.Context())
			if !ok {
				log.Warnf("[%s]: role not found in context", op)
				response.UnauthorizedJSON(w)
				return
			}

			isAllowed := false
			for _, allowedRole := range allowedRoles {
				if userRole == allowedRole {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				log.Warnf("[%s]: access denied for role %s", op, userRole)
				response.ForbiddenJSON(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
