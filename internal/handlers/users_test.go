package handlers

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/rshafikov/gophermart/internal/core"
	"github.com/rshafikov/gophermart/internal/core/security"
	"github.com/rshafikov/gophermart/internal/models"
	"github.com/rshafikov/gophermart/internal/repository"
	"github.com/rshafikov/gophermart/internal/service"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUserHandler_Register(t *testing.T) {
	userRepo := repository.NewMockUserRepository()
	userServce := service.NewUserService(userRepo)
	jwtHanlder := &security.MockJWTHandler{}
	handler := NewUserHandler(userServce, jwtHanlder)
	apiUserRegisterPath := "/api/user/register"

	r := chi.NewRouter()
	r.Post(apiUserRegisterPath, handler.Register)
	ts := httptest.NewServer(r)
	defer ts.Close()

	type want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name string
		url  string
		body string
		want want
	}{
		{
			name: "register a user",
			url:  apiUserRegisterPath,
			body: `{"login":"user_1","password":"password"}`,
			want: want{
				code:        http.StatusOK,
				response:    `{"token":"fake-token user_1","token_type":"Bearer"}`,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "register same user",
			url:  apiUserRegisterPath,
			body: `{"login":"user_1","password":"password"}`,
			want: want{
				code:        http.StatusConflict,
				response:    `login is not available`,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "register with empty login",
			url:  apiUserRegisterPath,
			body: `{"login":"","password":"password"}`,
			want: want{
				code:        http.StatusBadRequest,
				response:    `invalid login`,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "register without login",
			url:  apiUserRegisterPath,
			body: `{"password":"password"}`,
			want: want{
				code:        http.StatusBadRequest,
				response:    `invalid login`,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "register without password",
			url:  apiUserRegisterPath,
			body: `{"login":"login"}`,
			want: want{
				code:        http.StatusBadRequest,
				response:    `invalid password`,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	var notCompress bool
	client := core.NewHTTPClient(ts.URL, notCompress)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, body := client.JSONRequest(t, http.MethodPost, test.url, test.body)
			defer response.Body.Close()

			assert.Equal(t, test.want.code, response.StatusCode)
			assert.Equal(t, test.want.response, strings.Trim(body, "\n"))
			assert.Equal(t, test.want.contentType, response.Header.Get("Content-Type"))
		})
	}

}

func TestUserHandler_Login(t *testing.T) {
	userRepo := repository.NewMockUserRepository()
	userServce := service.NewUserService(userRepo)
	jwtHanlder := &security.MockJWTHandler{}
	handler := NewUserHandler(userServce, jwtHanlder)
	apiUserLoginPath := "/api/user/login"

	r := chi.NewRouter()
	r.Post(apiUserLoginPath, handler.Login)
	ts := httptest.NewServer(r)
	defer ts.Close()

	testUser1 := models.User{Login: "user_1", Password: "password1"}
	testUser2 := models.User{Login: "user_2", Password: "password2"}
	ctx := context.TODO()

	_ = userRepo.CreateUser(ctx, &testUser1)
	_ = userRepo.CreateUser(ctx, &testUser2)

	type want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name string
		url  string
		body string
		want want
	}{
		{
			name: "login user1",
			url:  apiUserLoginPath,
			body: `{"login":"user_1","password":"password1"}`,
			want: want{
				code:        http.StatusOK,
				response:    `{"token":"fake-token user_1","token_type":"Bearer"}`,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "login same user",
			url:  apiUserLoginPath,
			body: `{"login":"user_1","password":"password1"}`,
			want: want{
				code:        http.StatusOK,
				response:    `{"token":"fake-token user_1","token_type":"Bearer"}`,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "login another user",
			url:  apiUserLoginPath,
			body: `{"login":"user_2","password":"password2"}`,
			want: want{
				code:        http.StatusOK,
				response:    `{"token":"fake-token user_2","token_type":"Bearer"}`,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "register with wrong password",
			url:  apiUserLoginPath,
			body: `{"login":"user_1","password":"password"}`,
			want: want{
				code:        http.StatusUnauthorized,
				response:    `password mismatch`,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	var notCompress bool
	client := core.NewHTTPClient(ts.URL, notCompress)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, body := client.JSONRequest(t, http.MethodPost, test.url, test.body)
			defer response.Body.Close()

			assert.Equal(t, test.want.code, response.StatusCode)
			assert.Equal(t, test.want.response, strings.Trim(body, "\n"))
			assert.Equal(t, test.want.contentType, response.Header.Get("Content-Type"))
		})
	}

}
