package http

import (
	"net/http"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/microservices/core/auth/dto"
	"github.com/RBS-Team/Okoshki/pkg/response"
	"github.com/gorilla/csrf"
)

// GetCSRFToken обрабатывает запрос на получение CSRF токена.
// @Summary Получить CSRF токен
// @Description Возвращает новый CSRF токен для аутентифицированного пользователя. Токен должен быть включен в последующие запросы, изменяющие состояние (POST, PUT, DELETE и т.д.), либо в заголовок запроса, либо в поле формы.
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} dto.CsrfResponse "CSRF токен успешно сгенерирован"
// @Failure 401 {object} response.ErrorResponse "Не авторизован - отсутствует или недействительный токен аутентификации"
// @Failure 500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router /csrf-token [get]
func (h *AuthHandler) GetCSRFToken(w http.ResponseWriter, r *http.Request) {
	const op = "handler.user.GetCSRFToken"
	log := middleware.LoggerFromContext(r.Context())
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Errorf("[%s]: %s", op, "Не удалось получить userID из контекста")
	}
	//Теперь иди в pkg и придумывай там функцию генерации
	token := csrf.Token(r)

	log.Debugf("[%s]: successfully generated csrf token for user %s", op, userID)

	response.JSON(w, http.StatusOK, dto.CsrfResponse{Token: token})
}
