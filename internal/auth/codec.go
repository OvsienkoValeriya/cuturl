package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	cookieName = "auth_token"
)

var secret []byte
var ErrInvalidMAC = errors.New("invalid mac")

func Init(secretStr string) {
	secret = []byte(secretStr)
}

func createHMAC(data string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(data))
	return base64.URLEncoding.EncodeToString(mac.Sum(nil))
}

func validateToken(token string) (string, error) {
	parts := strings.Split(token, "|")
	if len(parts) != 2 {
		return "", ErrInvalidMAC
	}
	userID, sig := parts[0], parts[1]
	expectedSig := createHMAC(userID)
	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return "", ErrInvalidMAC
	}
	return userID, nil
}

func GenerateToken() string {
	userID := uuid.NewString()
	return userID
}

func SetAuthCookie(w http.ResponseWriter, userID string) {
	token := userID + "|" + createHMAC(userID)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(365 * 24 * time.Hour),
	})
}

func GetUserIDFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", err
	}
	return validateToken(cookie.Value)
}
