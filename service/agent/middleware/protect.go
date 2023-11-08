package middleware

import (
	"context"
	"net/http"
	"tinyvm/internal"
)

type ProtectHandlerFunc func(ctx context.Context) http.HandlerFunc

func Protect(ctx context.Context, next ProtectHandlerFunc) http.HandlerFunc {
	as := ctx.Value(internal.CTX_SECRET).(string)
	return func(w http.ResponseWriter, r *http.Request) {
		secret := r.Header.Get("Secret")
		if secret != as {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next(ctx)(w, r)
	}
}
