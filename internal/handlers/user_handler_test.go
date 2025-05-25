package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/rshafikov/gophermart/internal/core"
	"github.com/rshafikov/gophermart/internal/core/security"
	"github.com/rshafikov/gophermart/internal/mocks"
	"github.com/rshafikov/gophermart/internal/models"
	"github.com/rshafikov/gophermart/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUserHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserService(ctrl)
	mockJWTHandler := mocks.NewMockJWTHandler(ctrl)

	handler := NewUserHandler(mockUserService, mockJWTHandler)
	apiUserRegisterPath := "/api/user/register"

	r := chi.NewRouter()
	r.Post(apiUserRegisterPath, handler.Register)
	ts := httptest.NewServer(r)
	defer ts.Close()

	type want struct {
		code     int
		response string
		cType    string
		token    string
	}

	tests := []struct {
		name       string
		url        string
		body       string
		want       want
		setupMocks func()
	}{
		{
			name: "register user_1:password",
			url:  apiUserRegisterPath,
			body: `{"login":"user_1","password":"password"}`,
			want: want{
				code:     http.StatusOK,
				response: "",
				cType:    "",
				token:    `fake-token user_1`,
			},
			setupMocks: func() {
				mockUserService.EXPECT().Register(gomock.Any(), "user_1", "password").Return(nil)
				mockJWTHandler.EXPECT().GenerateJWT("user_1").Return(&security.JWTToken{Token: "fake-token user_1"}, nil)
			},
		},
		{
			name: "register a user_2:Zz123456!1",
			url:  apiUserRegisterPath,
			body: `{"login":"user_2","password":"Zz123456!1"}`,
			want: want{
				code:     http.StatusOK,
				response: "",
				cType:    "",
				token:    `fake-token user_2`,
			},
			setupMocks: func() {
				mockUserService.EXPECT().Register(gomock.Any(), "user_2", "Zz123456!1").Return(nil)
				mockJWTHandler.EXPECT().GenerateJWT("user_2").Return(&security.JWTToken{Token: "fake-token user_2"}, nil)
			},
		},
		{
			name: "register same user",
			url:  apiUserRegisterPath,
			body: `{"login":"user_1","password":"password"}`,
			want: want{
				code:     http.StatusConflict,
				response: `login is not available`,
				cType:    "text/plain; charset=utf-8",
				token:    "",
			},
			setupMocks: func() {
				mockUserService.EXPECT().Register(gomock.Any(), "user_1", "password").Return(service.ErrUserAlreadyExists)
			},
		},
		{
			name: "register with empty login",
			url:  apiUserRegisterPath,
			body: `{"login":"","password":"password"}`,
			want: want{
				code:     http.StatusBadRequest,
				response: `invalid login`,
				cType:    "text/plain; charset=utf-8",
				token:    "",
			},
		},
		{
			name: "register without login",
			url:  apiUserRegisterPath,
			body: `{"password":"password"}`,
			want: want{
				code:     http.StatusBadRequest,
				response: `invalid login`,
				cType:    "text/plain; charset=utf-8",
				token:    "",
			},
		},
		{
			name: "register without password",
			url:  apiUserRegisterPath,
			body: `{"login":"login"}`,
			want: want{
				code:     http.StatusBadRequest,
				response: `too short password`,
				cType:    "text/plain; charset=utf-8",
				token:    "",
			},
		},
	}

	var notCompress bool
	client := core.NewHTTPClient(ts.URL, notCompress)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMocks != nil {
				test.setupMocks()
			}

			resp, b := client.JSONRequest(t, http.MethodPost, test.url, test.body)
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode)
			assert.Equal(t, test.want.cType, resp.Header.Get("Content-Type"))
			assert.Equal(t, test.want.response, strings.Trim(b, "\n"))
			assert.Equal(t, test.want.token, resp.Header.Get("Authorization"))
		})
	}

}

func TestUserHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserService(ctrl)
	mockJWTHandler := mocks.NewMockJWTHandler(ctrl)
	handler := NewUserHandler(mockUserService, mockJWTHandler)
	apiUserLoginPath := "/api/user/login"

	r := chi.NewRouter()
	r.Post(apiUserLoginPath, handler.Login)
	ts := httptest.NewServer(r)
	defer ts.Close()

	testUser1 := models.User{Login: "user_1", Password: "password1"}
	testUser2 := models.User{Login: "user_2", Password: "password2"}

	type want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name       string
		url        string
		body       string
		want       want
		setupMocks func()
	}{
		{
			name: "login user1",
			url:  apiUserLoginPath,
			body: `{"login":"user_1","password":"password1"}`,
			want: want{
				code:        http.StatusOK,
				response:    `{"token":"fake-token user_1","token_type":"Bearer","expires_at":"0001-01-01T00:00:00Z"}`,
				contentType: "application/json; charset=utf-8",
			},
			setupMocks: func() {
				mockUserService.EXPECT().Login(gomock.Any(), "user_1", "password1").Return(&testUser1, nil)
				mockJWTHandler.EXPECT().GenerateJWT("user_1").Return(
					&security.JWTToken{Token: "fake-token user_1", TokenType: security.TokenType},
					nil,
				)
			},
		},
		{
			name: "login same user",
			url:  apiUserLoginPath,
			body: `{"login":"user_1","password":"password1"}`,
			want: want{
				code:        http.StatusOK,
				response:    `{"token":"fake-token user_1","token_type":"Bearer","expires_at":"0001-01-01T00:00:00Z"}`,
				contentType: "application/json; charset=utf-8",
			},
			setupMocks: func() {
				mockUserService.EXPECT().Login(gomock.Any(), "user_1", "password1").Return(&testUser1, nil)
				mockJWTHandler.EXPECT().GenerateJWT("user_1").Return(
					&security.JWTToken{Token: "fake-token user_1", TokenType: security.TokenType},
					nil,
				)
			},
		},
		{
			name: "login another user",
			url:  apiUserLoginPath,
			body: `{"login":"user_2","password":"password2"}`,
			want: want{
				code:        http.StatusOK,
				response:    `{"token":"fake-token user_2","token_type":"Bearer","expires_at":"0001-01-01T00:00:00Z"}`,
				contentType: "application/json; charset=utf-8",
			},
			setupMocks: func() {
				mockUserService.EXPECT().Login(gomock.Any(), "user_2", "password2").Return(&testUser2, nil)
				mockJWTHandler.EXPECT().GenerateJWT("user_2").Return(
					&security.JWTToken{Token: "fake-token user_2", TokenType: security.TokenType}, nil)
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
			setupMocks: func() {
				mockUserService.EXPECT().Login(gomock.Any(), "user_1", "password").Return(nil, service.ErrPasswordMismatch)
			},
		},
	}

	var notCompress bool
	client := core.NewHTTPClient(ts.URL, notCompress)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMocks != nil {
				test.setupMocks()
			}
			response, body := client.JSONRequest(t, http.MethodPost, test.url, test.body)
			defer response.Body.Close()

			assert.Equal(t, test.want.code, response.StatusCode)
			assert.Equal(t, test.want.response, strings.Trim(body, "\n"))
			assert.Equal(t, test.want.contentType, response.Header.Get("Content-Type"))
		})
	}

}
