package app

import (
	"cuturl/internal/config"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrigUrlHandler(t *testing.T) {
	config.Init()
	type want struct {
		code        int
		contentType string
		checkBody   func(t *testing.T, body string)
	}

	tests := []struct {
		name   string
		body   string
		method string
		want   want
	}{
		{
			name:   "positive test - shorten url",
			body:   "https://practicum.yandex.ru/",
			method: http.MethodPost,
			want: want{
				code:        http.StatusCreated,
				contentType: "text/plain",
				checkBody: func(t *testing.T, body string) {
					assert.True(t, strings.HasPrefix(body, "http://localhost:8080/"))
					assert.Equal(t, 8, len(strings.TrimPrefix(body, "http://localhost:8080/")))
				},
			},
		},
		{
			name:   "empty body",
			body:   "",
			method: http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				checkBody: func(t *testing.T, body string) {
					assert.Contains(t, body, "invalid request body")
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Post("/", OrigUrlHandler)
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()
			bodyBytes, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			body := string(bodyBytes)

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			tt.want.checkBody(t, body)
		})
	}
}

func TestShortUrlHandler(t *testing.T) {
	shortID := "SRHQaQLO"
	originalURL := "https://practicum.yandex.ru/"
	shortToOriginal[shortID] = originalURL

	type want struct {
		code      int
		location  string
		bodyCheck func(t *testing.T, body string)
	}

	tests := []struct {
		name   string
		id     string
		method string
		want   want
	}{
		{
			name:   "positive - redirect works",
			id:     shortID,
			method: http.MethodGet,
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: originalURL,
				bodyCheck: func(t *testing.T, body string) {
					assert.Empty(t, body)
				},
			},
		},
		{
			name:   "not existing id",
			id:     "notfound",
			method: http.MethodGet,
			want: want{
				code:     http.StatusBadRequest,
				location: "",
				bodyCheck: func(t *testing.T, body string) {
					assert.Contains(t, body, "short url not found")
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Get("/{id}", ShortUrlHandler)
			req := httptest.NewRequest(tt.method, "/"+tt.id, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			body := string(bodyBytes)

			assert.Equal(t, tt.want.code, resp.StatusCode)
			if tt.want.location != "" {
				assert.Equal(t, tt.want.location, resp.Header.Get("Location"))
			}
			tt.want.bodyCheck(t, body)
		})
	}
}
