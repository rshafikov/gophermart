package handlers

import (
	"bytes"
	"context"
	"errors"
	"github.com/rshafikov/gophermart/internal/core/security"
	"github.com/rshafikov/gophermart/internal/models"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

type MapUserRepository struct {
	DB map[string]*models.User
}

func NewMapUserRepository() models.UserRepository {
	return &MapUserRepository{DB: make(map[string]*models.User)}
}

func (m *MapUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	user.Password, _ = security.HashPassword(user.Password)
	m.DB[user.Login] = user
	return nil
}

func (m *MapUserRepository) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	user, ok := m.DB[login]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MapUserRepository) Clear() {
	m.DB = make(map[string]*models.User)
}

type MockJWTGenerator struct{}

func (m *MockJWTGenerator) GenerateJWT(login string) (*security.JWTToken, error) {
	return &security.JWTToken{Token: "fake-token " + login, TokenType: security.TokenType}, nil
}

func (m *MockJWTGenerator) ParseJWT(token string) (*security.TokenPayload, error) {
	return nil, nil
}

type HTTPClient struct {
	Client  *http.Client
	BaseURL string
}

func NewHTTPClient(baseURL string, isCompress bool) *HTTPClient {
	if isCompress {
		return &HTTPClient{
			Client:  http.DefaultClient,
			BaseURL: baseURL,
		}
	}

	return &HTTPClient{
		Client: &http.Client{
			Transport: &http.Transport{
				DisableCompression: true,
			},
		},
		BaseURL: baseURL,
	}
}

func (c *HTTPClient) URLRequest(t *testing.T, method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, c.BaseURL+path, nil)
	require.NoError(t, err)

	resp, err := c.Client.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func (c *HTTPClient) JSONRequest(t *testing.T, method, path, reqBody string) (*http.Response, string) {
	req, err := http.NewRequest(method, c.BaseURL+path, bytes.NewBuffer([]byte(reqBody)))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Client.Do(req)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(body)
}
