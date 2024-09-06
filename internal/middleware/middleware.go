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

		claims := Claims{}
		token, err := jwt.ParseWithClaims(bearerToken[1], claims, func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				http.Error(w, "Invalid token signature", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		if !token.Valid {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		r.WithContext(jwt.NewContext())
		r = r.WithContext(r.Context(), token)
		next(w, r)
		// token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
		// 	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		// 		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		// 	}
		// 	return []byte(m.config.JWT.Secret), nil
		// })

		// if err != nil {
		// 	http.Error(w, "Invalid token", http.StatusUnauthorized)
		// 	return
		// }

		// if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// 	ctx := context.WithValue(r.Context(), "user", claims)
		// 	next.ServeHTTP(w, r.WithContext(ctx))
		// } else {
		// 	http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		// }
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

func getUserIDFromContext(ctx context.Context) int {
	user := ctx.Value("user").(jwt.MapClaims)
	return int(user["id"].(float64))
}
