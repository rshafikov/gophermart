package middlewares

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/rshafikov/gophermart/internal/core"
	"github.com/rshafikov/gophermart/internal/core/contextkeys"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/core/security"
	"github.com/rshafikov/gophermart/internal/models"
	"github.com/rshafikov/gophermart/internal/repository"
	"github.com/rshafikov/gophermart/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	userService := service.NewUserService(repository.NewMockUserRepository())
	jwtHandler := &security.MockJWTHandler{}
	authMW := Authenticater(jwtHandler, userService)

	testUser := models.User{Login: "user_1", Password: "password"}
	err := userService.Register(context.TODO(), testUser.Login, testUser.Password)
	require.NoError(t, err)

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
		name  string
		want  want
		token string
	}{
		{
			name:  "test with valid token",
			token: "fake-token user_1",
			want: want{
				code:     http.StatusOK,
				response: "user_1",
			},
		},
		{
			name:  "test with invalid token",
			token: "wrong-fake-token user_1",
			want: want{
				code:     http.StatusUnauthorized,
				response: "unauthorized",
			},
		},
	}

	var notCompress bool
	c := core.NewHTTPClient(ts.URL, notCompress)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
}
