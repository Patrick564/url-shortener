package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Patrick564/url-shortener-backend/api/controllers"
	"github.com/Patrick564/url-shortener-backend/internal/models"
	"github.com/Patrick564/url-shortener-backend/utils"
)

type mockUrlModel struct{}

func (m *mockUrlModel) All() ([]models.Url, error) {
	var urls []models.Url

	urls = append(urls, models.Url{ShortUrl: "ID_1", OriginalUrl: "https://www.example-url-1.com"})
	urls = append(urls, models.Url{ShortUrl: "ID_2", OriginalUrl: "https://www.example-url-2.dev"})
	urls = append(urls, models.Url{ShortUrl: "ID_3", OriginalUrl: "https://www.example-url-3.com"})

	return urls, nil
}

func (m *mockUrlModel) Add(url string) (models.Url, error) {
	if url != "https://www.example.com" {
		return models.Url{}, utils.ErrEmptyBody
	}

	return models.Url{ShortUrl: "ID_1", OriginalUrl: "https://www.example.com"}, nil
}

type allRouteResponse struct {
	Error error        `json:"error"`
	Urls  []models.Url `json:"urls"`
}

type addRouteResponse struct {
	Error error      `json:"error"`
	Url   models.Url `json:"url"`
}

func TestAllRoute(t *testing.T) {
	env := &controllers.Env{Urls: &mockUrlModel{}}
	router := SetupRouter(env)

	tests := []struct {
		name      string
		wantCode  int
		wantError error
		wantBody  []byte
		res       allRouteResponse
	}{
		{
			name:      "Returns without errors",
			wantCode:  http.StatusOK,
			wantError: nil,
			res: allRouteResponse{
				Error: nil,
				Urls: []models.Url{
					{ShortUrl: "ID_1", OriginalUrl: "https://www.example-url-1.com"},
					{ShortUrl: "ID_2", OriginalUrl: "https://www.example-url-2.dev"},
					{ShortUrl: "ID_3", OriginalUrl: "https://www.example-url-3.com"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/all", nil)
			router.ServeHTTP(w, req)

			wantBody, _ := json.Marshal(tt.res)

			utils.AssertStatusCode(t, tt.wantCode, w.Code)
			utils.AssertResponseBody(t, wantBody, w.Body.String())
		})
	}
}

func TestAddRoute(t *testing.T) {
	env := &controllers.Env{Urls: &mockUrlModel{}}
	router := SetupRouter(env)

	tests := []struct {
		name      string
		body      io.Reader
		wantCode  int
		wantError error
		res       addRouteResponse
	}{
		{
			name:      "Returns without errors",
			body:      bytes.NewBuffer([]byte("{\"url\": \"https://www.example.com\" }")),
			wantCode:  http.StatusOK,
			wantError: nil,
			res: addRouteResponse{
				Error: nil,
				Url:   models.Url{ShortUrl: "ID_1", OriginalUrl: "https://www.example.com"},
			},
		},
		{
			name:      "Returns error with empty body",
			body:      bytes.NewBuffer([]byte("")),
			wantCode:  http.StatusBadRequest,
			wantError: utils.ErrEmptyBody,
			res: addRouteResponse{
				Error: utils.ErrEmptyBody,
				Url:   models.Url{},
			},
		},
		{
			name:      "Returns error with incorrect url",
			body:      bytes.NewBuffer([]byte("{\"url\": \"ejemplo-bad-url:4040\" }")),
			wantCode:  http.StatusBadRequest,
			wantError: utils.ErrInvalidUrl,
			res: addRouteResponse{
				Error: utils.ErrInvalidUrl,
				Url:   models.Url{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/add", tt.body)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			wantBody, _ := json.Marshal(tt.res)

			utils.AssertError(t, tt.wantError, tt.res.Error)
			utils.AssertStatusCode(t, tt.wantCode, w.Code)
			utils.AssertResponseBody(t, wantBody, w.Body.String())
		})
	}
}
