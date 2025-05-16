package handlers

import (
	"encoding/json"
	"errors"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/core/security"
	"github.com/rshafikov/gophermart/internal/schemas"
	"github.com/rshafikov/gophermart/internal/service"
	"go.uber.org/zap"
	"net/http"
)

type UserHandler struct {
	UserService *service.UserService
	JWT         security.JWTHandler
}

func NewUserHandler(userService *service.UserService, jwtService security.JWTHandler) *UserHandler {
	return &UserHandler{UserService: userService, JWT: jwtService}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var reqUser schemas.UserCreate
	if err := json.NewDecoder(r.Body).Decode(&reqUser); err != nil {
		logger.L.Debug("unable to decode request body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.ValidateUserCredentials(reqUser.Login, reqUser.Password); err != nil {
		logger.L.Debug("invalid credentials", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.UserService.Register(ctx, reqUser.Login, reqUser.Password); err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			logger.L.Debug("user already exists", zap.String("login", reqUser.Login))
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		logger.L.Debug("unable to register user", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := h.JWT.GenerateJWT(reqUser.Login)
	if err != nil {
		logger.L.Debug("unable to generate JWT", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		logger.L.Debug("unable to encode JWT", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err = w.Write(tokenBytes)
	if err != nil {
		logger.L.Debug("unable to write JWT", zap.Error(err))
		return
	}
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var reqUser schemas.UserCreate
	if err := json.NewDecoder(r.Body).Decode(&reqUser); err != nil {
		logger.L.Debug("unable to decode request body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.UserService.Login(ctx, reqUser.Login, reqUser.Password)
	if err != nil {
		logger.L.Debug("unable to login with given credentials", zap.Error(err))
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	token, err := h.JWT.GenerateJWT(user.Login)
	if err != nil {
		logger.L.Debug("unable to generate JWT", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		logger.L.Debug("unable to encode JWT", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(tokenBytes)
	if err != nil {
		logger.L.Debug("unable to write JWT", zap.Error(err))
		return
	}
}

func (h *UserHandler) ValidateUserCredentials(login string, password string) error {
	if !security.IsLoginValid(login) {
		return errors.New("invalid login")
	}
	if !security.IsPasswordValid(password) {
		return errors.New("invalid password")
	}
	return nil
}
