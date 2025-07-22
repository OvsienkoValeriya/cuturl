package middleware

import (
	"context"
	"net/http"

	"cuturl/internal/auth"
)

type CtxKey string

const UserIDKey CtxKey = "userID"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil || userID == "" {
			userID := auth.GenerateToken()
			auth.SetAuthCookie(w, userID)
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
