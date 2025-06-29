package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rshafikov/gophermart/internal/core"
	"github.com/rshafikov/gophermart/internal/core/security"
	"github.com/rshafikov/gophermart/internal/middlewares"
	"github.com/rshafikov/gophermart/internal/mocks"
	"github.com/rshafikov/gophermart/internal/models"
	"github.com/rshafikov/gophermart/internal/service"
	"github.com/rshafikov/gophermart/internal/workerpool"
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

func TestOrderHandler_GetOrders(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	mockUserService := mocks.NewMockuserService(mockController)
	mockJWTHandler := mocks.NewMockJWTHandler(mockController)
	authMiddleware := middlewares.Authenticater(mockJWTHandler, mockUserService)

	mockOrderService := mocks.NewMockorderService(mockController)
	mockBalanceService := mocks.NewMockbalanceService(mockController)
	mockClient := mocks.NewMockClient(mockController)
	wp := workerpool.NewWorkerPool(1, mockClient, mockOrderService, mockBalanceService)

	testingHandler := NewOrderHandler(mockOrderService, wp)
	apiURL := "/api/user/orders"

	r := chi.NewRouter()
	r.Use(authMiddleware)
	r.Get(apiURL, testingHandler.GetOrders)
	ts := httptest.NewServer(r)
	defer ts.Close()

	type want struct {
		code     int
		response string
		cType    string
	}
	tests := []struct {
		name       string
		want       want
		token      string
		setupMocks func()
	}{
		{
			name: "user1 recieves orders",
			want: want{
				code:     http.StatusOK,
				response: `[{"number":"12345678903","status":"PROCESSING","uploaded_at":"2020-12-10T15:12:01Z"},{"number":"346436439","status":"INVALID","uploaded_at":"2020-12-09T16:09:53Z"}]`,
				cType:    "application/json; charset=utf-8",
			},
			token: "Bearer valid-token",
			setupMocks: func() {
				u := models.User{ID: 1, Login: "user1"}
				orderProcessing := models.Order{
					NumeralID: "12345678903",
					UserID:    u.ID,
					Status:    models.OrderStatusProcessing,
					Accrual:   0,
					CreatedAt: time.Date(2020, 12, 10, 15, 12, 01, 0, time.UTC),
				}
				orderInvalid := models.Order{
					NumeralID: "346436439",
					UserID:    u.ID,
					Status:    models.OrderStatusInvalid,
					Accrual:   0,
					CreatedAt: time.Date(2020, 12, 9, 16, 9, 53, 0, time.UTC),
				}

				mockJWTHandler.EXPECT().ParseJWT("valid-token").
					Return(&security.TokenPayload{RegisteredClaims: jwt.RegisteredClaims{Subject: u.Login}}, nil)

				mockUserService.EXPECT().GetByLogin(gomock.Any(), u.Login).
					Return(&u, nil)

				mockOrderService.EXPECT().GetOrders(gomock.Any(), u.ID).
					Return([]*models.Order{&orderProcessing, &orderInvalid}, nil)
			},
		},
		{
			name: "user2 recieves no orders",
			want: want{
				code:     http.StatusNoContent,
				response: ``,
				cType:    "",
			},
			token: "Bearer valid-token",
			setupMocks: func() {
				u := models.User{ID: 2, Login: "user2"}

				mockJWTHandler.EXPECT().ParseJWT("valid-token").
					Return(&security.TokenPayload{RegisteredClaims: jwt.RegisteredClaims{Subject: u.Login}}, nil)

				mockUserService.EXPECT().GetByLogin(gomock.Any(), u.Login).
					Return(&u, nil)

				mockOrderService.EXPECT().GetOrders(gomock.Any(), u.ID).
					Return([]*models.Order{}, nil)
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

func TestOrderHandler_CreateOrder(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	mockUserService := mocks.NewMockuserService(mockController)
	mockJWTHandler := mocks.NewMockJWTHandler(mockController)
	authMiddleware := middlewares.Authenticater(mockJWTHandler, mockUserService)

	mockOrderService := mocks.NewMockorderService(mockController)
	mockBalanceService := mocks.NewMockbalanceService(mockController)
	mockClient := mocks.NewMockClient(mockController)
	wp := workerpool.NewWorkerPool(1, mockClient, mockOrderService, mockBalanceService)

	testingHandler := NewOrderHandler(mockOrderService, wp)
	apiURL := "/api/user/orders"

	r := chi.NewRouter()
	r.Use(authMiddleware)
	r.Post(apiURL, testingHandler.CreateOrder)
	ts := httptest.NewServer(r)
	defer ts.Close()

	type want struct {
		code     int
		response string
		cType    string
	}
	tests := []struct {
		name       string
		want       want
		token      string
		body       string
		setupMocks func()
	}{
		{
			name: "invalid order number",
			want: want{
				code:     http.StatusUnprocessableEntity,
				response: `invalid order number`,
				cType:    "text/plain; charset=utf-8",
			},
			body:  "123456789",
			token: "Bearer valid-token",
			setupMocks: func() {
				u := models.User{ID: 1, Login: "user1"}

				mockJWTHandler.EXPECT().ParseJWT("valid-token").
					Return(&security.TokenPayload{RegisteredClaims: jwt.RegisteredClaims{Subject: u.Login}}, nil)

				mockUserService.EXPECT().GetByLogin(gomock.Any(), u.Login).
					Return(&u, nil)
			},
		},
		{
			name: "successful order creation",
			want: want{
				code:     http.StatusAccepted,
				response: "",
				cType:    "",
			},
			body:  "12345678903",
			token: "Bearer valid-token",
			setupMocks: func() {
				u := models.User{ID: 2, Login: "user2"}

				newOrder := models.Order{
					NumeralID: "12345678903",
					UserID:    u.ID,
					Status:    models.OrderStatusNew,
				}

				mockJWTHandler.EXPECT().ParseJWT("valid-token").
					Return(&security.TokenPayload{RegisteredClaims: jwt.RegisteredClaims{Subject: u.Login}}, nil)

				mockUserService.EXPECT().GetByLogin(gomock.Any(), u.Login).
					Return(&u, nil)

				mockOrderService.EXPECT().CreateOrderIfNotExists(gomock.Any(), &newOrder).
					Return(nil)
			},
		},
		{
			name: "duplicate order",
			want: want{
				code:     http.StatusOK,
				response: "",
				cType:    "",
			},
			body:  "12345678903",
			token: "Bearer valid-token",
			setupMocks: func() {
				u := models.User{ID: 1, Login: "user2"}

				sameOrder := models.Order{
					NumeralID: "12345678903",
					UserID:    u.ID,
					Status:    models.OrderStatusNew,
				}

				mockJWTHandler.EXPECT().ParseJWT("valid-token").
					Return(&security.TokenPayload{RegisteredClaims: jwt.RegisteredClaims{Subject: u.Login}}, nil)

				mockUserService.EXPECT().GetByLogin(gomock.Any(), u.Login).
					Return(&u, nil)

				mockOrderService.EXPECT().CreateOrderIfNotExists(gomock.Any(), &sameOrder).
					Return(service.ErrOrderAlreadyLoaded)
			},
		},
		{
			name: "order loaded by someone else",
			want: want{
				code:     http.StatusConflict,
				response: "order was loaded by another user",
				cType:    "text/plain; charset=utf-8",
			},
			body:  "12345678903",
			token: "Bearer valid-token",
			setupMocks: func() {
				u := models.User{ID: 3, Login: "user3"}

				sameOrder := models.Order{
					NumeralID: "12345678903",
					UserID:    u.ID,
					Status:    models.OrderStatusNew,
				}

				mockJWTHandler.EXPECT().ParseJWT("valid-token").
					Return(&security.TokenPayload{RegisteredClaims: jwt.RegisteredClaims{Subject: u.Login}}, nil)

				mockUserService.EXPECT().GetByLogin(gomock.Any(), u.Login).
					Return(&u, nil)

				mockOrderService.EXPECT().CreateOrderIfNotExists(gomock.Any(), &sameOrder).
					Return(service.ErrOrderLoadedBySomeone)
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

			req, err := http.NewRequest(http.MethodPost, ts.URL+apiURL, strings.NewReader(test.body))
			require.NoError(t, err)

			req.Header.Set("Authorization", test.token)
			req.Header.Set("Content-Type", "text/plain")

			resp, err := c.Client.Do(req)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode)
			assert.Equal(t, test.want.cType, resp.Header.Get("Content-Type"))
			assert.Equal(t, test.want.response, strings.Trim(string(body), "\n"))
		})
	}
}
