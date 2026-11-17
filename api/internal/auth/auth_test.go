package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pquerna/otp/totp"

	"github.com/batuhangozegu/metron/api/internal/cryptobox"
)

// fakeStore models atomic store operations without a database.
type fakeStore struct {
	users    map[string]User
	sessions map[string]Session
}

func newFake() *fakeStore {
	return &fakeStore{users: map[string]User{}, sessions: map[string]Session{}}
}
func (f *fakeStore) Register(_ context.Context, email, hash string, s Session) (User, error) {
	for _, u := range f.users {
		if u.Email == email {
			return User{}, ErrConflict
		}
	}
	u := User{ID: fmt.Sprint(len(f.users) + 1), Email: email, PasswordHash: hash, CreatedAt: s.CreatedAt}
	f.users[u.ID] = u
	s.UserID = u.ID
	f.sessions[string(s.Hash)] = s
	return u, nil
}
func (f *fakeStore) UserByEmail(_ context.Context, email string) (User, error) {
	for _, u := range f.users {
		if u.Email == email {
			return u, nil
		}
	}
	return User{}, ErrUnauthorized
}
func (f *fakeStore) UserByID(_ context.Context, id string) (User, error) {
	u, ok := f.users[id]
	if !ok {
		return User{}, ErrUnauthorized
	}
	return u, nil
}
func (f *fakeStore) InsertSession(_ context.Context, s Session) error {
	if f.users[s.UserID].TwoFactorEnabled != s.Pending {
		return ErrState
	}
	f.sessions[string(s.Hash)] = s
	return nil
}
func (f *fakeStore) ResolveSession(_ context.Context, hash []byte, pending bool, now time.Time) (User, error) {
	s, ok := f.sessions[string(hash)]
	if !ok || s.Pending != pending || !s.ExpiresAt.After(now) {
		return User{}, ErrUnauthorized
	}
	if !pending {
		s.ExpiresAt = now.Add(SessionTTL)
		f.sessions[string(hash)] = s
	}
	return f.users[s.UserID], nil
}
func (f *fakeStore) DeleteSession(_ context.Context, hash []byte) error {
	delete(f.sessions, string(hash))
	return nil
}
func (f *fakeStore) SetSecret(_ context.Context, id string, secret []byte, _ time.Time) error {
	u := f.users[id]
	if u.TwoFactorEnabled {
		return ErrState
	}
	u.Secret = secret
	u.LastStep = nil
	f.users[id] = u
	return nil
}
func (f *fakeStore) ApplyTOTP(_ context.Context, c TOTPChange) error {
	u := f.users[c.UserID]
	if !bytes.Equal(u.Secret, c.Secret) || (u.LastStep != nil && *u.LastStep >= c.Step) || u.TwoFactorEnabled != (c.Action != "enable") {
		return ErrCode
	}
	if c.Action == "login" {
		pending, ok := f.sessions[string(c.CurrentHash)]
		if !ok || !pending.Pending || pending.UserID != c.UserID || !pending.ExpiresAt.After(c.Now) {
			return ErrUnauthorized
		}
		delete(f.sessions, string(c.CurrentHash))
		f.sessions[string(c.NewSession.Hash)] = c.NewSession
	} else {
		for key, s := range f.sessions {
			if s.UserID == c.UserID && key != string(c.CurrentHash) {
				delete(f.sessions, key)
			}
		}
	}
	u.TwoFactorEnabled = c.Action != "disable"
	u.LastStep = &c.Step
	if !u.TwoFactorEnabled {
		u.Secret = nil
	}
	f.users[u.ID] = u
	return nil
}

func testService(t *testing.T) (*Service, *fakeStore, *time.Time) {
	t.Helper()
	box, err := cryptobox.New(base64.StdEncoding.EncodeToString(make([]byte, 32)))
	if err != nil {
		t.Fatal(err)
	}
	f := newFake()
	s, err := New(f, box)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	s.Now = func() time.Time { return now }
	return s, f, &now
}

func TestRegistrationValidation(t *testing.T) {
	for _, tc := range []struct {
		name, email, password, want string
		valid                       bool
	}{
		{"normalized", " Test@Example.COM ", "twelve-chars!", "test@example.com", true},
		{"minimum", "a@example.com", strings.Repeat("a", 12), "a@example.com", true},
		{"maximum", "a@example.com", strings.Repeat("é", 128), "a@example.com", true},
		{"short", "a@example.com", "12345678901", "", false},
		{"long", "a@example.com", strings.Repeat("a", 129), "", false},
		{"invalid email", "invalid", "twelve-chars!", "", false},
		{"display name", "User <a@example.com>", "twelve-chars!", "", false},
		{"empty", "", "twelve-chars!", "", false},
		{"newline", "a@example.com\r\nBcc:x@example.com", "twelve-chars!", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateRegistration(tc.email, tc.password)
			if (err == nil) != tc.valid || got != tc.want {
				t.Fatalf("got %q, %v", got, err)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	s, f, _ := testService(t)
	user, token, err := s.Register(context.Background(), " User@example.com ", "correct password!", "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}
	if f.users[user.ID].PasswordHash == "correct password!" {
		t.Fatal("plaintext password stored")
	}
	if _, ok := f.sessions[token]; ok {
		t.Fatal("raw token stored")
	}
	for _, tc := range []struct {
		name, email, password string
		want                  error
	}{
		{"success", "USER@EXAMPLE.COM", "correct password!", nil},
		{"wrong password", "user@example.com", "wrong password!", ErrInvalid},
		{"unknown email", "missing@example.com", "correct password!", ErrInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			u, raw, pending, err := s.Login(context.Background(), tc.email, tc.password, "127.0.0.1", "test")
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
			if err == nil && (u.ID != user.ID || len(raw) != 43 || pending) {
				t.Fatal("invalid login result")
			}
		})
	}
	_, _, err = s.Register(context.Background(), "user@EXAMPLE.com", "correct password!", "127.0.0.1", "test")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate: %v", err)
	}
}

func TestSessionExpiry(t *testing.T) {
	for _, tc := range []struct {
		name    string
		pending bool
		elapsed time.Duration
		valid   bool
	}{
		{"full active", false, SessionTTL - time.Second, true},
		{"full boundary", false, SessionTTL, false},
		{"full expired", false, SessionTTL + time.Second, false},
		{"pending active", true, PendingTTL - time.Second, true},
		{"pending boundary", true, PendingTTL, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, f, now := testService(t)
			f.users["1"] = User{ID: "1", TwoFactorEnabled: tc.pending}
			token, session, err := s.session("1", "127.0.0.1", "test", tc.pending)
			if err != nil {
				t.Fatal(err)
			}
			f.sessions[string(session.Hash)] = session
			*now = now.Add(tc.elapsed)
			_, err = s.Authenticate(context.Background(), token, tc.pending)
			if (err == nil) != tc.valid {
				t.Fatalf("got %v", err)
			}
			if tc.valid {
				expiry := f.sessions[string(session.Hash)].ExpiresAt
				want := now.Add(SessionTTL)
				if tc.pending {
					want = session.ExpiresAt
				}
				if !expiry.Equal(want) {
					t.Fatalf("expiry %v, want %v", expiry, want)
				}
			}
			if _, err = s.Authenticate(context.Background(), token, !tc.pending); err == nil {
				t.Fatal("wrong session kind accepted")
			}
		})
	}
}

func TestTwoFactorFlow(t *testing.T) {
	ctx := context.Background()
	s, f, now := testService(t)
	user, original, err := s.Register(ctx, "user@example.com", "correct password!", "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}
	_, other, _, err := s.Login(ctx, user.Email, "correct password!", "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}
	secret, url, err := s.Setup(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, "otpauth://totp/") || bytes.Contains(f.users[user.ID].Secret, []byte(secret)) {
		t.Fatal("secret must be encrypted")
	}
	code := func() string {
		t.Helper()
		c, err := totp.GenerateCode(secret, *now)
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	if err = s.ChangeTOTP(ctx, user.ID, original, "enable", "bad", "", "127.0.0.1"); !errors.Is(err, ErrCode) {
		t.Fatal(err)
	}
	if err = s.ChangeTOTP(ctx, user.ID, original, "enable", code(), "", "127.0.0.1"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Authenticate(ctx, other, false); err == nil {
		t.Fatal("old session survived enable")
	}
	if _, _, err = s.Setup(ctx, user.ID); !errors.Is(err, ErrState) {
		t.Fatal("enabled secret replaced")
	}
	_, pending, isPending, err := s.Login(ctx, user.Email, "correct password!", "127.0.0.1", "test")
	if err != nil || !isPending {
		t.Fatalf("pending login: %v", err)
	}
	if _, err = s.Authenticate(ctx, pending, false); err == nil {
		t.Fatal("pending token authorized")
	}
	if _, _, err = s.CompleteLogin(ctx, pending, code(), "127.0.0.1", "test"); !errors.Is(err, ErrCode) {
		t.Fatalf("enable code replay: %v", err)
	}
	*now = now.Add(30 * time.Second)
	_, full, err := s.CompleteLogin(ctx, pending, code(), "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Authenticate(ctx, full, false); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Authenticate(ctx, pending, true); err == nil {
		t.Fatal("pending token not consumed")
	}
	*now = now.Add(30 * time.Second)
	if err = s.ChangeTOTP(ctx, user.ID, full, "disable", code(), "wrong password", "127.0.0.1"); !errors.Is(err, ErrInvalid) {
		t.Fatal("disabled without password")
	}
	if err = s.ChangeTOTP(ctx, user.ID, full, "disable", code(), "correct password!", "127.0.0.1"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Authenticate(ctx, original, false); err == nil {
		t.Fatal("old session survived disable")
	}
	_, _, isPending, err = s.Login(ctx, user.Email, "correct password!", "127.0.0.1", "test")
	if err != nil || isPending {
		t.Fatalf("login after disable: %v", err)
	}
}

func TestTOTPWindows(t *testing.T) {
	s, _, now := testService(t)
	secret := "JBSWY3DPEHPK3PXP"
	encrypted, err := s.Box.Seal([]byte(secret), "totp:1")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name          string
		offset        int64
		reused, valid bool
	}{
		{"current", 0, false, true},
		{"previous", -1, false, true},
		{"next", 1, false, true},
		{"too old", -2, false, false},
		{"too far ahead", 2, false, false},
		{"current reused", 0, true, false},
		{"previous reused", -1, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			counter := now.Unix()/30 + tc.offset
			code, err := totp.GenerateCode(secret, time.Unix(counter*30, 0))
			if err != nil {
				t.Fatal(err)
			}
			user := User{ID: "1", Secret: encrypted}
			if tc.reused {
				user.LastStep = &counter
			}
			got, err := s.step(user, code)
			if (err == nil) != tc.valid {
				t.Fatalf("step=%d err=%v", got, err)
			}
			if tc.valid && got != counter {
				t.Fatalf("got %d want %d", got, counter)
			}
		})
	}
}

func TestCookies(t *testing.T) {
	now := time.Now().UTC()
	for _, secure := range []bool{false, true} {
		for _, name := range []string{CookieName, PendingCookieName} {
			for _, token := range []string{"random-token", ""} {
				t.Run(fmt.Sprintf("secure=%v/%s/clear=%v", secure, name, token == ""), func(t *testing.T) {
					h := HTTP{Secure: secure}
					c := h.Cookie(name, token, now)
					if c.HttpOnly != true || c.Secure != secure || c.SameSite != http.SameSiteLaxMode || c.Path != "/" || c.Domain != "" {
						t.Fatalf("flags: %+v", c)
					}
					ttl := SessionTTL
					if name == PendingCookieName {
						ttl = PendingTTL
					}
					if token == "" {
						if c.MaxAge != -1 || !c.Expires.Before(now) {
							t.Fatal("not cleared")
						}
					} else if c.MaxAge != int(ttl.Seconds()) || !c.Expires.Equal(now.Add(ttl)) {
						t.Fatal("wrong expiry")
					}
				})
			}
		}
	}
}

func TestHTTPPolicies(t *testing.T) {
	s, _, _ := testService(t)
	for _, tc := range []struct {
		name, path, method, body, contentType string
		key                                   bool
		want                                  int
	}{
		{"missing key", "/v1/auth/register", "POST", `{}`, "application/json", false, 503},
		{"unauthorized me", "/v1/me", "GET", "", "", true, 401},
		{"form rejected", "/v1/auth/login", "POST", "email=a", "application/x-www-form-urlencoded", true, 415},
		{"unknown field", "/v1/auth/register", "POST", `{"extra":1}`, "application/json", true, 400},
		{"trailing JSON", "/v1/auth/register", "POST", `{} {}`, "application/json", true, 400},
		{"numeric code", "/v1/auth/login/2fa", "POST", `{"code":123456}`, "application/json", true, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := &Service{Store: s.Store, Box: s.Box, Now: s.Now, dummyHash: s.dummyHash}
			if !tc.key {
				service.Box = nil
			}
			r := chi.NewRouter()
			HTTP{Service: service}.Mount(r, nil)
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", tc.contentType)
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)
			if res.Code != tc.want || !strings.Contains(res.Header().Get("Content-Type"), "application/json") {
				t.Fatalf("%d: %s", res.Code, res.Body.String())
			}
		})
	}
}

func TestLimiter(t *testing.T) {
	var l Limiter
	now := time.Now()
	for i := 0; i < 10; i++ {
		if !l.Allow(now, "email:a", "ip:1") {
			t.Fatal("limited too early")
		}
	}
	if l.Allow(now, "email:a", "ip:2") || l.Allow(now, "email:b", "ip:1") {
		t.Fatal("both email and IP must be limited")
	}
	if !l.Allow(now.Add(15*time.Minute), "email:a", "ip:1") {
		t.Fatal("window did not reset")
	}
}

func TestPasswordHash(t *testing.T) {
	hash, err := HashPassword("correct password!")
	if err != nil {
		t.Fatal(err)
	}
	other, err := HashPassword("correct password!")
	if err != nil {
		t.Fatal(err)
	}
	if hash == other {
		t.Fatal("salt reused")
	}
	for _, tc := range []struct {
		hash, password string
		valid          bool
	}{
		{hash, "correct password!", true}, {hash, "incorrect password", false}, {"bad hash", "correct password!", false},
		{strings.Replace(hash, "m=65536", "m=999999999", 1), "correct password!", false},
	} {
		if got := VerifyPassword(tc.hash, tc.password); got != tc.valid {
			t.Fatalf("verification got %v", got)
		}
	}
}
