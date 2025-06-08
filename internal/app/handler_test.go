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

func TestOrigURLHandler(t *testing.T) {
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
					assert.True(t, strings.HasPrefix(body, config.Get().BaseURL))
					assert.Equal(t, 8, len(strings.TrimPrefix(body, config.Get().BaseURL)))
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
			u := NewURLShortener()

			r := chi.NewRouter()
			r.Post("/", http.HandlerFunc(u.OrigURLHandler))

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
