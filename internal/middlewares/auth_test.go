package middlewares

import (
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rshafikov/gophermart/internal/core"
	"github.com/rshafikov/gophermart/internal/core/contextkeys"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/core/security"
	"github.com/rshafikov/gophermart/internal/mocks"
	"github.com/rshafikov/gophermart/internal/models"
	"github.com/rshafikov/gophermart/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(contextkeys.UserKey).(*models.User)
	if !ok {
		logger.L.Debug("user not found in context")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, err := w.Write([]byte(u.Login))
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

func TestAuthenticater(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserService(ctrl)
	mockJWTHandler := mocks.NewMockJWTHandler(ctrl)
	authMW := Authenticater(mockJWTHandler, mockUserService)

	testUser := models.User{Login: "user_1", Password: "password"}

	r := chi.NewRouter()
	r.Use(authMW)
	r.Get("/", testHandler)
	ts := httptest.NewServer(r)
	defer ts.Close()

	type want struct {
		code     int
		response string
	}

	tests := []struct {
		name       string
		want       want
		token      string
		setupMocks func()
	}{
		{
			name:  "test with valid token",
			token: "fake-token user_1",
			want: want{
				code:     http.StatusOK,
				response: "user_1",
			},
			setupMocks: func() {
				mockJWTHandler.EXPECT().ParseJWT("fake-token user_1").Return(
					&security.TokenPayload{RegisteredClaims: jwt.RegisteredClaims{Subject: "user_1"}}, nil)
				mockUserService.EXPECT().GetByLogin(gomock.Any(), "user_1").Return(&testUser, nil)
			},
		},
		{
			name:  "test with invalid token",
			token: "wrong-fake-token user_1",
			want: want{
				code:     http.StatusUnauthorized,
				response: "unauthorized",
			},
			setupMocks: func() {
				mockJWTHandler.EXPECT().ParseJWT("wrong-fake-token user_1").Return(
					&security.TokenPayload{}, security.ErrTokenInvalid)
			},
		},
		{
			name:  "test token when user not found",
			token: "fake-token user_1",
			want: want{
				code:     http.StatusUnauthorized,
				response: "unauthorized",
			},
			setupMocks: func() {
				mockJWTHandler.EXPECT().ParseJWT("fake-token user_1").Return(
					&security.TokenPayload{RegisteredClaims: jwt.RegisteredClaims{Subject: "user_1"}}, nil)
				mockUserService.EXPECT().GetByLogin(gomock.Any(), "user_1").Return(nil, service.ErrUserNotFound)
			},
		},
	}

	var notCompress bool
	c := core.NewHTTPClient(ts.URL, notCompress)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMocks != nil {
				test.setupMocks()
			}

			req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
			require.NoError(t, err)

			req.Header.Set("Authorization", "Bearer "+test.token)
			resp, err := c.Client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, test.want.code, resp.StatusCode)
			assert.Equal(t, test.want.response, strings.Trim(string(body), "\n"))
		})
	}

	t.Run("test with empty Authorization header", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
		require.NoError(t, err)

		resp, err := c.Client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
