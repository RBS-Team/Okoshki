package http

//go:generate easyjson $GOFILE

import (
	"net/http"
	"time"

	"github.com/RBS-Team/Okoshki/internal/middleware"

	"github.com/RBS-Team/Okoshki/microservices/core/auth/dto"
	"github.com/RBS-Team/Okoshki/pkg/response"

	easyjson "github.com/mailru/easyjson"
)

const (
	sessionTokenCookie = "session_token"
)

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	const op = "handler.Register"
	defer r.Body.Close()

	log := middleware.LoggerFromContext(r.Context())

	var req dto.RegisterRequest

	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		log.Errorf("[%s]: Invalid request body: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	user, err := h.service.Register(r.Context(), dto.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		log.Errorf("[%s]: Service error: %v", op, err)
		h.handleError(w, err)
		return
	}

	token, err := h.jwtManager.NewToken(user.ID, user.Role)
	if err != nil {
		log.Errorf("[%s]: Failed to generate token: %v", op, err)
		response.InternalErrorJSON(w)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionTokenCookie,
		Value:    token,
		Expires:  time.Now().Add(h.jwtManager.GetTTL()),
		HttpOnly: true,
		Path:     "/",
	})

	log.Infof("[%s]: User registered successfully: %s", op, user.ID)
	response.JSON(w, http.StatusCreated, dto.RegisterResponse{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	const op = "handler.Login"
	defer r.Body.Close()

	log := middleware.LoggerFromContext(r.Context())

	var req dto.LoginRequest

	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		log.Errorf("[%s]: Invalid request body: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	
	user, err := h.service.Login(r.Context(), dto.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		log.Errorf("[%s]: Service error: %v", op, err)
		h.handleError(w, err)
		return
	}

	token, err := h.jwtManager.NewToken(user.ID,user.Role)
	if err != nil {
		log.Errorf("[%s]: Failed to generate token: %v", op, err)
		response.InternalErrorJSON(w)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionTokenCookie,
		Value:    token,
		Expires:  time.Now().Add(h.jwtManager.GetTTL()),
		HttpOnly: true,
		Path:     "/",
	})

	log.Infof("[%s]: User login successfully: %s", op, user.ID)
	response.JSON(w, http.StatusOK, dto.LoginResponse{ID: user.ID})
}

// func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
// 	const op = "handler.Logout"
// 	log := middleware.LoggerFromContext(r.Context())

// 	http.SetCookie(w, &http.Cookie{
// 		Name:     sessionTokenCookie,
// 		Value:    "",
// 		Expires:  time.Now().Add(-time.Hour),
// 		HttpOnly: true,
// 		Path:     "/",
// 	})

// 	log.Infof("[%s]: User logout successful", op)
// 	response.JSON(w, http.StatusOK, logoutResponse{Status: "ok"})
// }
