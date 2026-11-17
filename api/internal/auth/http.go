package auth

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/batuhangozegu/metron/api/internal/httpjson"
)

const CookieName = "metron_session"
const PendingCookieName = "metron_pending"

type userKey struct{}

func UserFromContext(ctx context.Context) User { u, _ := ctx.Value(userKey{}).(User); return u }

type HTTP struct {
	Service *Service
	Secure  bool
}

func (h HTTP) Cookie(name, token string, now time.Time) *http.Cookie {
	ttl := SessionTTL
	if name == PendingCookieName {
		ttl = PendingTTL
	}
	c := &http.Cookie{Name: name, Value: token, Path: "/", HttpOnly: true, Secure: h.Secure, SameSite: http.SameSiteLaxMode, Expires: now.Add(ttl), MaxAge: int(ttl.Seconds())}
	if token == "" {
		c.MaxAge = -1
		c.Expires = time.Unix(1, 0).UTC()
	}
	return c
}

func tokenFrom(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.String()
	}
	return "0.0.0.0"
}

func (h HTTP) Ready(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.Service == nil || h.Service.Box == nil {
			httpjson.Error(w, 503, "authentication unavailable: configure APP_ENCRYPTION_KEY (32 bytes, base64)")
			return
		}
		if h.Service.Store == nil {
			httpjson.Error(w, 503, "authentication unavailable: configure DATABASE_URL and run migrations")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h HTTP) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := tokenFrom(r, CookieName)
		user, err := h.Service.Authenticate(r.Context(), token, false)
		if err != nil {
			h.fail(w, err)
			return
		}
		http.SetCookie(w, h.Cookie(CookieName, token, h.Service.Now()))
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey{}, user)))
	})
}

// Mount also accepts protected module routes so account handlers share the
// exact same readiness, JSON-only and session policy as /v1/me.
func (h HTTP) Mount(r chi.Router, protected func(chi.Router)) {
	r.Group(func(r chi.Router) {
		r.Use(h.Ready, httpjson.JSONOnly)
		r.Post("/v1/auth/register", h.register)
		r.Post("/v1/auth/login", h.login)
		r.Post("/v1/auth/login/2fa", h.complete)
		r.Post("/v1/auth/logout", h.logout)
		r.Group(func(r chi.Router) {
			r.Use(h.Require)
			r.Get("/v1/me", func(w http.ResponseWriter, r *http.Request) { httpjson.Write(w, 200, UserFromContext(r.Context())) })
			r.Post("/v1/auth/2fa/setup", h.setup)
			r.Post("/v1/auth/2fa/enable", h.enable)
			r.Post("/v1/auth/2fa/disable", h.disable)
			if protected != nil {
				protected(r)
			}
		})
	})
}

func (h HTTP) fail(w http.ResponseWriter, err error) {
	var validation ValidationError
	switch {
	case errors.As(err, &validation):
		httpjson.Error(w, 400, err.Error())
	case errors.Is(err, ErrUnauthorized), errors.Is(err, ErrInvalid):
		httpjson.Error(w, 401, err.Error())
	case errors.Is(err, ErrCode):
		httpjson.Error(w, 401, err.Error())
	case errors.Is(err, ErrConflict), errors.Is(err, ErrState):
		httpjson.Error(w, 409, err.Error())
	case errors.Is(err, ErrLimited):
		w.Header().Set("Retry-After", "900")
		httpjson.Error(w, 429, err.Error())
	default:
		httpjson.Error(w, 503, "authentication temporarily unavailable")
	}
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h HTTP) signedIn(w http.ResponseWriter, user User, token string, status int) {
	http.SetCookie(w, h.Cookie(CookieName, token, h.Service.Now()))
	http.SetCookie(w, h.Cookie(PendingCookieName, "", h.Service.Now()))
	httpjson.Write(w, status, user)
}

func (h HTTP) register(w http.ResponseWriter, r *http.Request) {
	var body credentials
	if !httpjson.Decode(w, r, &body) {
		return
	}
	user, token, err := h.Service.Register(r.Context(), body.Email, body.Password, clientIP(r), r.UserAgent())
	if err != nil {
		h.fail(w, err)
		return
	}
	h.signedIn(w, user, token, 201)
}

func (h HTTP) login(w http.ResponseWriter, r *http.Request) {
	var body credentials
	if !httpjson.Decode(w, r, &body) {
		return
	}
	user, token, pending, err := h.Service.Login(r.Context(), body.Email, body.Password, clientIP(r), r.UserAgent())
	if err != nil {
		h.fail(w, err)
		return
	}
	if pending {
		http.SetCookie(w, h.Cookie(PendingCookieName, token, h.Service.Now()))
		http.SetCookie(w, h.Cookie(CookieName, "", h.Service.Now()))
		httpjson.Write(w, 200, map[string]bool{"two_factor_required": true})
		return
	}
	h.signedIn(w, user, token, 200)
}

func (h HTTP) complete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code string `json:"code"`
	}
	if !httpjson.Decode(w, r, &body) {
		return
	}
	user, token, err := h.Service.CompleteLogin(r.Context(), tokenFrom(r, PendingCookieName), body.Code, clientIP(r), r.UserAgent())
	if err != nil {
		h.fail(w, err)
		return
	}
	h.signedIn(w, user, token, 200)
}

func (h HTTP) logout(w http.ResponseWriter, r *http.Request) {
	for _, name := range []string{CookieName, PendingCookieName} {
		token := tokenFrom(r, name)
		if token != "" {
			if err := h.Service.Store.DeleteSession(r.Context(), TokenHash(token)); err != nil {
				h.fail(w, err)
				return
			}
		}
		http.SetCookie(w, h.Cookie(name, "", h.Service.Now()))
	}
	httpjson.Write(w, 200, map[string]bool{"ok": true})
}

func (h HTTP) setup(w http.ResponseWriter, r *http.Request) {
	var body struct{}
	if !httpjson.Decode(w, r, &body) {
		return
	}
	secret, url, err := h.Service.Setup(r.Context(), UserFromContext(r.Context()).ID)
	if err != nil {
		h.fail(w, err)
		return
	}
	httpjson.Write(w, 200, map[string]string{"secret": secret, "otpauth_url": url})
}

func (h HTTP) enable(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code string `json:"code"`
	}
	if !httpjson.Decode(w, r, &body) {
		return
	}
	err := h.Service.ChangeTOTP(r.Context(), UserFromContext(r.Context()).ID, tokenFrom(r, CookieName), "enable", body.Code, "", clientIP(r))
	if err != nil {
		h.fail(w, err)
		return
	}
	httpjson.Write(w, 200, map[string]bool{"two_factor_enabled": true})
}

func (h HTTP) disable(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code     string `json:"code"`
		Password string `json:"password"`
	}
	if !httpjson.Decode(w, r, &body) {
		return
	}
	err := h.Service.ChangeTOTP(r.Context(), UserFromContext(r.Context()).ID, tokenFrom(r, CookieName), "disable", body.Code, body.Password, clientIP(r))
	if err != nil {
		h.fail(w, err)
		return
	}
	httpjson.Write(w, 200, map[string]bool{"two_factor_enabled": false})
}
