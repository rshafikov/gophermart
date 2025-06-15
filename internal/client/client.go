package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/schemas"
	"go.uber.org/zap"
	"io"
	"net/http"
)

var ErrInternalServerError = errors.New("accrual Internal Server Error")
var ErrTooManyRequests = errors.New("too many requests")
var ErrNoContent = errors.New("no content")
var ErrInnerError = errors.New("inner error")

type Client interface {
	GetOrderStatus(ctx context.Context, number string) (*schemas.AccrualOrder, int, error)
}

type accrualClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAccrualClient(baseURL string) Client {
	return &accrualClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (r *accrualClient) GetOrderStatus(ctx context.Context, number string) (*schemas.AccrualOrder, int, error) {
	url := fmt.Sprintf("%s/api/orders/%s", r.baseURL, number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.L.Debug("Error creating request", zap.Error(err))
		return nil, -1, ErrInnerError
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, -1, ErrInnerError
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, -1, ErrInnerError
	}

	switch statusCode {
	case http.StatusOK:
		var order schemas.AccrualOrder
		if err := json.Unmarshal(body, &order); err != nil {
			return nil, http.StatusInternalServerError, err
		}
		return &order, statusCode, nil

	case http.StatusNoContent:
		return nil, statusCode, ErrNoContent

	case http.StatusTooManyRequests:
		return nil, statusCode, ErrTooManyRequests

	case http.StatusInternalServerError:
		return nil, statusCode, ErrInternalServerError

	default:
		logger.L.Warn("unexpected status code", zap.Int("status_code", statusCode))
		return nil, statusCode, ErrInternalServerError
	}
}
