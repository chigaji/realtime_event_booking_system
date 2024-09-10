package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chigaji/realtime_event_booking_system/internal/config"
	"github.com/go-redis/redis/v8"
	jwt "github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type contextKey string

const userClaimKey contextKey = "userClaim"

var jwtKey = []byte("secreteKey")

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type Middleware struct {
	redis  *redis.Client
	logger *zap.Logger
	config *config.Config
}

func NewMiddleware(redis *redis.Client, logger *zap.Logger, config *config.Config) *Middleware {
	return &Middleware{
		redis:  redis,
		logger: logger,
		config: config,
	}

}

func GenerateJWTToken(username string) (string, error) {

	expirationTime := time.Now().Add(time.Minute * 60)

	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (m *Middleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		claims := &Claims{}
		tokenString := bearerToken[1]

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtKey, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				http.Error(w, "Invalid token signature", http.StatusUnauthorized)
				return
			}
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !token.Valid {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		// fmt.Println("claims --->", token.Claims)

		ctx := context.WithValue(r.Context(), userClaimKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (m *Middleware) RateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := getUserIDFromContext(ctx)
		key := fmt.Sprintf("rate_limit:%d", userID)

		count, err := m.redis.Incr(ctx, key).Result()
		if err != nil {
			m.logger.Error("Failed to increment rate limit counter", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if count == 1 {
			m.redis.Expire(ctx, key, time.Minute)
		}

		if count > m.config.RateLimit.RequestsPerMinute {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func getUserIDFromContext(ctx context.Context) jwt.Claims {
	user, ok := ctx.Value(userClaimKey).(*Claims)
	if !ok {
		return nil
	}
	return user
}
