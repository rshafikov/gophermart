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
	"time"
)

const requestTimeout = 3 * time.Second

var ErrAccrualServerFailure = errors.New("accrual Internal Server Error")
var ErrTooManyRequests = errors.New("too many requests")
var ErrNoContent = errors.New("no content")
var ErrInnerError = errors.New("inner error")

type Client interface {
	GetOrderStatus(ctx context.Context, number string) (*schemas.AccrualOrder, error)
}

type accrualClient struct {
	baseURL        string
	httpClient     *http.Client
	requestTimeout time.Duration
}

func NewAccrualClient(baseURL string) Client {
	return &accrualClient{
		baseURL:        "http://" + baseURL,
		httpClient:     &http.Client{},
		requestTimeout: requestTimeout,
	}
}

func (r *accrualClient) GetOrderStatus(ctx context.Context, number string) (*schemas.AccrualOrder, error) {
	url := fmt.Sprintf("%s/api/orders/%s", r.baseURL, number)

	requestCtx, cancel := context.WithTimeout(ctx, r.requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, url, nil)
	if err != nil {
		logger.L.Debug("error creating request", zap.Error(err))
		return nil, ErrInnerError
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			logger.L.Debug("request timed out", zap.Error(err))
			return nil, fmt.Errorf("request timed out: %w", err)
		}
		logger.L.Debug("error performing request", zap.Error(err))
		return nil, ErrInnerError
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrInnerError
	}
	logger.L.Info("response from accrual service", zap.Int("status_code", statusCode), zap.ByteString("body", body))

	switch statusCode {
	case http.StatusOK:
		var order schemas.AccrualOrder
		if err := json.Unmarshal(body, &order); err != nil {
			return nil, ErrInnerError
		}
		return &order, nil

	case http.StatusNoContent:
		return nil, ErrNoContent

	case http.StatusTooManyRequests:
		return nil, ErrTooManyRequests

	case http.StatusInternalServerError:
		return nil, ErrAccrualServerFailure

	default:
		logger.L.Debug("unexpected status code", zap.Int("status_code", statusCode))
		return nil, ErrAccrualServerFailure
	}
}
