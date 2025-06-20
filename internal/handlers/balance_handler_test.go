package handlers

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rshafikov/gophermart/internal/core"
	"github.com/rshafikov/gophermart/internal/core/security"
	"github.com/rshafikov/gophermart/internal/middlewares"
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
	"time"
)

func TestBalanceHandler_GetUserBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBalanceService := mocks.NewMockBalanceService(ctrl)
	mockUserService := mocks.NewMockUserService(ctrl)
	mockJWTHandler := mocks.NewMockJWTHandler(ctrl)
	authMiddleware := middlewares.Authenticater(mockJWTHandler, mockUserService)

	handler := NewBalanceHandler(mockBalanceService)
	apiURL := "/api/user/balance"

	r := chi.NewRouter()
	r.Use(authMiddleware)
	r.Get(apiURL, handler.GetUserBalance)
	ts := httptest.NewServer(r)
	defer ts.Close()

	type want struct {
		code     int
		response string
		cType    string
	}
	tests := []struct {
		name       string
		token      string
		want       want
		setupMocks func()
	}{
		{
			name: "user1 not authenticated to get balance",
			want: want{
				code:     http.StatusUnauthorized,
				response: `unauthorized`,
				cType:    "text/plain; charset=utf-8",
			},
			token: "Bearer invalid-token",
			setupMocks: func() {
				mockJWTHandler.EXPECT().
					ParseJWT("invalid-token").
					Return(&security.TokenPayload{}, security.ErrTokenInvalid)
			},
		},
		{
			name: "successful balance retrieval for user1",
			want: want{
				code:     http.StatusOK,
				response: `{"current":100.5,"withdrawn":20.25}`,
				cType:    "application/json; charset=utf-8",
			},
			token: "Bearer valid-token",
			setupMocks: func() {
				u := models.User{ID: 1, Login: "user1"}

				mockJWTHandler.EXPECT().
					ParseJWT("valid-token").
					Return(&security.TokenPayload{RegisteredClaims: jwt.RegisteredClaims{Subject: u.Login}}, nil)

				mockUserService.EXPECT().
					GetByLogin(gomock.Any(), u.Login).
					Return(&u, nil)

				mockBalanceService.EXPECT().
					GetUserBalance(gomock.Any(), u.ID).
					Return(&models.Balance{Current: 100.50, Withdrawn: 20.25}, nil)
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

			req, err := http.NewRequest(http.MethodGet, ts.URL+apiURL, nil)
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", test.token)
			resp, err := c.Client.Do(req)
			require.NoError(t, err)

			b, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode)
			assert.Equal(t, test.want.cType, resp.Header.Get("Content-Type"))
			assert.Equal(t, test.want.response, strings.Trim(string(b), "\n"))
		})
	}
}

func TestBalanceHandler_GetUserWithdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBalanceService := mocks.NewMockBalanceService(ctrl)
	mockUserService := mocks.NewMockUserService(ctrl)
	mockJWTHandler := mocks.NewMockJWTHandler(ctrl)
	authMiddleware := middlewares.Authenticater(mockJWTHandler, mockUserService)

	handler := NewBalanceHandler(mockBalanceService)
	apiURL := "/api/user/withdrawals"

	r := chi.NewRouter()
	r.Use(authMiddleware)
	r.Get(apiURL, handler.GetUserWithdrawals)
	ts := httptest.NewServer(r)
	defer ts.Close()

	type want struct {
		code     int
		response string
		cType    string
	}
	tests := []struct {
		name       string
		token      string
		want       want
		setupMocks func()
	}{
		{
			name: "get user1 withdrawals",
			want: want{
				code:     http.StatusOK,
				response: `[{"order":"2377225624","sum":500,"processed_at":"2020-12-09T16:09:57Z"},{"order":"3129","sum":1337.5,"processed_at":"2020-12-09T16:09:56Z"}]`,
				cType:    "application/json; charset=utf-8",
			},
			token: "Bearer valid-token",
			setupMocks: func() {
				u := models.User{ID: 1, Login: "user1"}
				wd1 := models.Wd{
					OrderNumeralID: "2377225624",
					Amount:         500,
					CreatedAt:      time.Date(2020, 12, 9, 16, 9, 57, 0, time.UTC)}
				wd2 := models.Wd{OrderNumeralID: "3129", Amount: 1337.5, CreatedAt: time.Date(2020, 12, 9, 16, 9, 56, 0, time.UTC)}

				mockJWTHandler.EXPECT().
					ParseJWT("valid-token").
					Return(&security.TokenPayload{RegisteredClaims: jwt.RegisteredClaims{Subject: u.Login}}, nil)

				mockUserService.EXPECT().
					GetByLogin(gomock.Any(), u.Login).
					Return(&u, nil)

				mockBalanceService.EXPECT().
					GetWithdrawalsByUser(gomock.Any(), u.ID).
					Return([]*models.Wd{&wd1, &wd2}, nil)
			},
		},
		{
			name: "get no withdrawals for user2",
			want: want{
				code:     http.StatusNoContent,
				response: ``,
				cType:    "application/json; charset=utf-8",
			},
			token: "Bearer valid-token",
			setupMocks: func() {
				u := models.User{ID: 2, Login: "user2"}

				mockJWTHandler.EXPECT().
					ParseJWT("valid-token").
					Return(&security.TokenPayload{RegisteredClaims: jwt.RegisteredClaims{Subject: u.Login}}, nil)

				mockUserService.EXPECT().
					GetByLogin(gomock.Any(), u.Login).
					Return(&u, nil)

				mockBalanceService.EXPECT().
					GetWithdrawalsByUser(gomock.Any(), u.ID).
					Return([]*models.Wd{}, nil)
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

			resp, body := c.JSONRequest(t, http.MethodGet, apiURL, "", test.token)
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode)
			assert.Equal(t, test.want.cType, resp.Header.Get("Content-Type"))
			assert.Equal(t, test.want.response, strings.Trim(body, "\n"))
		})
	}
}

func TestBalanceHandler_WithdrawFromBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBalanceService := mocks.NewMockBalanceService(ctrl)
	mockUserService := mocks.NewMockUserService(ctrl)
	mockJWTHandler := mocks.NewMockJWTHandler(ctrl)
	authMiddleware := middlewares.Authenticater(mockJWTHandler, mockUserService)

	handler := NewBalanceHandler(mockBalanceService)
	testURL := "/api/user/balance/withdraw"

	r := chi.NewRouter()
	r.Use(authMiddleware)
	r.Post(testURL, handler.WithdrawFromBalance)
	ts := httptest.NewServer(r)
	defer ts.Close()

	type want struct {
		code     int
		response string
		cType    string
	}

	tests := []struct {
		name        string
		want        want
		jsonReqBody string
		token       string
		setupMocks  func()
	}{
		{
			name: "withdraw from balance with empty json body",
			want: want{
				code:     http.StatusBadRequest,
				response: "invalid request format",
				cType:    "text/plain; charset=utf-8",
			},
			jsonReqBody: ``,
			token:       "Bearer valid-token",
			setupMocks: func() {
				u := models.User{ID: 1, Login: "qwsedrftyhiujko"}

				mockJWTHandler.EXPECT().ParseJWT("valid-token").
					Return(&security.TokenPayload{RegisteredClaims: jwt.RegisteredClaims{Subject: u.Login}}, nil)

				mockUserService.EXPECT().GetByLogin(gomock.Any(), u.Login).
					Return(&u, nil)
			},
		},
		{
			name: "withdraw from balance with wrong json body",
			want: want{
				code:     http.StatusBadRequest,
				response: "invalid request format",
				cType:    "text/plain; charset=utf-8",
			},
			jsonReqBody: `{"order": {"order_id": 1234131},"summa": 751}`,
			token:       "Bearer valid-token",
			setupMocks: func() {
				u := models.User{ID: 1, Login: "qwsedrftyhiujko"}

				mockJWTHandler.EXPECT().ParseJWT("valid-token").
					Return(&security.TokenPayload{RegisteredClaims: jwt.RegisteredClaims{Subject: u.Login}}, nil)

				mockUserService.EXPECT().GetByLogin(gomock.Any(), u.Login).
					Return(&u, nil)
			},
		},
		{
			name: "withdraw from balance with wrong order number in body",
			want: want{
				code:     http.StatusUnprocessableEntity,
				response: models.ErrWdOrderIDValidationFailed.Error(),
				cType:    "text/plain; charset=utf-8",
			},
			jsonReqBody: `{"order": "23","sum": 751}`,
			token:       "Bearer valid-token",
			setupMocks: func() {
				u := models.User{ID: 1, Login: "qwsedrftyhiujko"}

				mockJWTHandler.EXPECT().ParseJWT("valid-token").
					Return(&security.TokenPayload{RegisteredClaims: jwt.RegisteredClaims{Subject: u.Login}}, nil)

				mockUserService.EXPECT().GetByLogin(gomock.Any(), u.Login).
					Return(&u, nil)
			},
		},
		{
			name: "withdraw from balance with wrong amount in body",
			want: want{
				code:     http.StatusUnprocessableEntity,
				response: models.ErrWdAmountValidationFailed.Error(),
				cType:    "text/plain; charset=utf-8",
			},
			jsonReqBody: `{"order": "2377225624","sum": -751}`,
			token:       "Bearer valid-token",
			setupMocks: func() {
				u := models.User{ID: 1, Login: "qwsedrftyhiujko"}

				mockJWTHandler.EXPECT().ParseJWT("valid-token").
					Return(&security.TokenPayload{RegisteredClaims: jwt.RegisteredClaims{Subject: u.Login}}, nil)

				mockUserService.EXPECT().GetByLogin(gomock.Any(), u.Login).
					Return(&u, nil)
			},
		},
		{
			name: "withdraw from balance successfully",
			want: want{
				code:     http.StatusOK,
				response: "",
				cType:    "",
			},
			jsonReqBody: `{"order": "2377225624","sum": 751}`,
			token:       "Bearer valid-token",
			setupMocks: func() {
				u := models.User{ID: 1, Login: "qwsedrftyhiujko"}
				wd := models.Wd{OrderNumeralID: "2377225624", Amount: 751}
				userBalance := models.Balance{Current: 751.50, Withdrawn: 20.25}

				mockJWTHandler.EXPECT().ParseJWT("valid-token").
					Return(&security.TokenPayload{RegisteredClaims: jwt.RegisteredClaims{Subject: u.Login}}, nil)

				mockUserService.EXPECT().GetByLogin(gomock.Any(), u.Login).
					Return(&u, nil)

				mockBalanceService.EXPECT().GetUserBalance(gomock.Any(), u.ID).
					Return(&userBalance, nil)

				mockBalanceService.EXPECT().Withdraw(gomock.Any(), &userBalance, &wd).Return(nil)
			},
		},
		{
			name: "withdraw from balance with insufficient funds",
			want: want{
				code:     http.StatusPaymentRequired,
				response: "insufficient funds",
				cType:    "text/plain; charset=utf-8",
			},
			jsonReqBody: `{"order": "2377225624","sum": 1000}`,
			token:       "Bearer valid-token",
			setupMocks: func() {
				u := models.User{ID: 1, Login: "qwsedrftyhiujko"}
				wd := models.Wd{OrderNumeralID: "2377225624", Amount: 1000}
				userBalance := models.Balance{Current: 500.00, Withdrawn: 20.25}

				mockJWTHandler.EXPECT().ParseJWT("valid-token").
					Return(&security.TokenPayload{RegisteredClaims: jwt.RegisteredClaims{Subject: u.Login}}, nil)

				mockUserService.EXPECT().GetByLogin(gomock.Any(), u.Login).
					Return(&u, nil)

				mockBalanceService.EXPECT().GetUserBalance(gomock.Any(), u.ID).
					Return(&userBalance, nil)

				mockBalanceService.EXPECT().Withdraw(gomock.Any(), &userBalance, &wd).
					Return(service.ErrInsufficientFunds)
			},
		},
		{
			name: "withdraw from balance with server error",
			want: want{
				code:     http.StatusInternalServerError,
				response: "internal server error",
				cType:    "text/plain; charset=utf-8",
			},
			jsonReqBody: `{"order": "2377225624","sum": 751}`,
			token:       "Bearer valid-token",
			setupMocks: func() {
				u := models.User{ID: 1, Login: "qwsedrftyhiujko"}
				wd := models.Wd{OrderNumeralID: "2377225624", Amount: 751}
				userBalance := models.Balance{Current: 751.50, Withdrawn: 20.25}

				mockJWTHandler.EXPECT().ParseJWT("valid-token").
					Return(&security.TokenPayload{RegisteredClaims: jwt.RegisteredClaims{Subject: u.Login}}, nil)

				mockUserService.EXPECT().GetByLogin(gomock.Any(), u.Login).
					Return(&u, nil)

				mockBalanceService.EXPECT().GetUserBalance(gomock.Any(), u.ID).
					Return(&userBalance, nil)

				mockBalanceService.EXPECT().Withdraw(gomock.Any(), &userBalance, &wd).
					Return(errors.New("database error"))
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

			resp, body := c.JSONRequest(t, http.MethodPost, testURL, test.jsonReqBody, test.token)
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode)
			assert.Equal(t, test.want.cType, resp.Header.Get("Content-Type"))
			assert.Equal(t, test.want.response, strings.Trim(body, "\n"))
		})
	}

}
