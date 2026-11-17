// Package auth implements password, TOTP and opaque cookie sessions.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/batuhangozegu/metron/api/internal/cryptobox"
)

const SessionTTL = 30 * 24 * time.Hour
const PendingTTL = 5 * time.Minute

var (
	ErrInvalid      = errors.New("invalid email or password")
	ErrUnauthorized = errors.New("authentication required")
	ErrConflict     = errors.New("email already registered")
	ErrCode         = errors.New("invalid or already used two-factor code")
	ErrState        = errors.New("two-factor state changed; start again")
	ErrLimited      = errors.New("too many attempts; try again in 15 minutes")
)

type ValidationError string

func (e ValidationError) Error() string { return string(e) }

type User struct {
	ID               string    `json:"id"`
	Email            string    `json:"email"`
	TwoFactorEnabled bool      `json:"two_factor_enabled"`
	CreatedAt        time.Time `json:"created_at"`
	PasswordHash     string    `json:"-"`
	Secret           []byte    `json:"-"`
	LastStep         *int64    `json:"-"`
}

type Session struct {
	UserID               string
	Hash                 []byte
	Pending              bool
	CreatedAt, ExpiresAt time.Time
	UserAgent, IP        string
}

type TOTPChange struct {
	UserID      string
	Secret      []byte // compare-and-swap protects against concurrent setup/reset
	Step        int64
	Action      string // enable, disable, login
	CurrentHash []byte
	NewSession  Session // only for login; consumes CurrentHash pending session
	Now         time.Time
}

type Store interface {
	Register(context.Context, string, string, Session) (User, error)
	UserByEmail(context.Context, string) (User, error)
	UserByID(context.Context, string) (User, error)
	InsertSession(context.Context, Session) error
	// ResolveSession slides only full sessions and never revives expired ones.
	ResolveSession(context.Context, []byte, bool, time.Time) (User, error)
	DeleteSession(context.Context, []byte) error
	SetSecret(context.Context, string, []byte, time.Time) error
	// ApplyTOTP atomically consumes the counter and updates sessions/state.
	ApplyTOTP(context.Context, TOTPChange) error
}

type Service struct {
	Store     Store
	Box       *cryptobox.Box
	Now       func() time.Time
	Limiter   Limiter
	dummyHash string
}

func New(store Store, box *cryptobox.Box) (*Service, error) {
	dummy, err := HashPassword("dummy password for constant work")
	if err != nil {
		return nil, err
	}
	return &Service{Store: store, Box: box, Now: func() time.Time { return time.Now().UTC() }, dummyHash: dummy}, nil
}

func NormalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email || len(email) > 254 || !strings.Contains(email, "@") {
		return "", ValidationError("valid email required")
	}
	return email, nil
}

func ValidateRegistration(email, password string) (string, error) {
	normalized, err := NormalizeEmail(email)
	if err != nil {
		return "", err
	}
	if !utf8.ValidString(password) || utf8.RuneCountInString(password) < 12 || utf8.RuneCountInString(password) > 128 {
		return "", ValidationError("password must contain 12 to 128 characters")
	}
	return normalized, nil
}

func TokenHash(token string) []byte { h := sha256.Sum256([]byte(token)); return h[:] }

func (s *Service) session(userID, ip, agent string, pending bool) (string, Session, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", Session{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(bytes)
	ttl := SessionTTL
	if pending {
		ttl = PendingTTL
	}
	now := s.Now()
	return token, Session{UserID: userID, Hash: TokenHash(token), Pending: pending, CreatedAt: now, ExpiresAt: now.Add(ttl), IP: ip, UserAgent: agent}, nil
}

func (s *Service) Register(ctx context.Context, email, password, ip, agent string) (User, string, error) {
	email, err := ValidateRegistration(email, password)
	if err != nil {
		return User{}, "", err
	}
	if !s.Limiter.Allow(s.Now(), "ip:"+ip, "email:"+email) {
		return User{}, "", ErrLimited
	}
	hash, err := HashPassword(password)
	if err != nil {
		return User{}, "", err
	}
	token, session, err := s.session("", ip, agent, false)
	if err != nil {
		return User{}, "", err
	}
	user, err := s.Store.Register(ctx, email, hash, session)
	return user, token, err
}

func (s *Service) Login(ctx context.Context, email, password, ip, agent string) (User, string, bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !s.Limiter.Allow(s.Now(), "ip:"+ip, "email:"+email) {
		return User{}, "", false, ErrLimited
	}
	user, err := s.Store.UserByEmail(ctx, email)
	hash := s.dummyHash
	if err == nil {
		hash = user.PasswordHash
	}
	// Unknown users perform the same Argon2id work as a wrong password.
	valid := VerifyPassword(hash, password)
	if err != nil && !errors.Is(err, ErrUnauthorized) {
		return User{}, "", false, err
	}
	if err != nil || !valid || utf8.RuneCountInString(password) > 128 {
		return User{}, "", false, ErrInvalid
	}
	token, session, err := s.session(user.ID, ip, agent, user.TwoFactorEnabled)
	if err != nil {
		return User{}, "", false, err
	}
	err = s.Store.InsertSession(ctx, session)
	return user, token, session.Pending, err
}

func (s *Service) Authenticate(ctx context.Context, token string, pending bool) (User, error) {
	if len(token) != 43 {
		return User{}, ErrUnauthorized
	}
	return s.Store.ResolveSession(ctx, TokenHash(token), pending, s.Now())
}

func (s *Service) Setup(ctx context.Context, id string) (string, string, error) {
	user, err := s.Store.UserByID(ctx, id)
	if err != nil {
		return "", "", err
	}
	if user.TwoFactorEnabled {
		return "", "", ErrState
	}
	key, err := totp.Generate(totp.GenerateOpts{Issuer: "Metron", AccountName: user.Email, Period: 30, SecretSize: 20, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
	if err != nil {
		return "", "", err
	}
	encrypted, err := s.Box.Seal([]byte(key.Secret()), "totp:"+id)
	if err != nil {
		return "", "", err
	}
	if err = s.Store.SetSecret(ctx, id, encrypted, s.Now()); err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

// Check each counter explicitly so its value can be consumed atomically in DB.
func (s *Service) step(user User, code string) (int64, error) {
	if len(code) != 6 {
		return 0, ErrCode
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			return 0, ErrCode
		}
	}
	secret, err := s.Box.Open(user.Secret, "totp:"+user.ID)
	if err != nil {
		return 0, err
	}
	counter := s.Now().Unix() / 30
	for _, offset := range []int64{0, -1, 1} {
		step := counter + offset
		valid, err := totp.ValidateCustom(code, string(secret), time.Unix(step*30, 0), totp.ValidateOpts{Period: 30, Skew: 0, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
		if err != nil {
			return 0, err
		}
		if valid && (user.LastStep == nil || step > *user.LastStep) {
			return step, nil
		}
	}
	return 0, ErrCode
}

func (s *Service) ChangeTOTP(ctx context.Context, userID, token, action, code, password, ip string) error {
	user, err := s.Store.UserByID(ctx, userID)
	if err != nil {
		return err
	}
	if !s.Limiter.Allow(s.Now(), "2fa-ip:"+ip, "2fa-user:"+user.Email) {
		return ErrLimited
	}
	if (action == "enable" && user.TwoFactorEnabled) || (action == "disable" && !user.TwoFactorEnabled) || len(user.Secret) == 0 {
		return ErrState
	}
	if action == "disable" && !VerifyPassword(user.PasswordHash, password) {
		return ErrInvalid
	}
	step, err := s.step(user, code)
	if err != nil {
		return err
	}
	return s.Store.ApplyTOTP(ctx, TOTPChange{UserID: user.ID, Secret: user.Secret, Step: step, Action: action, CurrentHash: TokenHash(token), Now: s.Now()})
}

func (s *Service) CompleteLogin(ctx context.Context, pendingToken, code, ip, agent string) (User, string, error) {
	if !s.Limiter.Allow(s.Now(), "2fa-ip:"+ip) {
		return User{}, "", ErrLimited
	}
	user, err := s.Authenticate(ctx, pendingToken, true)
	if err != nil {
		return User{}, "", err
	}
	if !s.Limiter.Allow(s.Now(), "2fa-user:"+user.Email) {
		return User{}, "", ErrLimited
	}
	if !user.TwoFactorEnabled {
		return User{}, "", ErrState
	}
	step, err := s.step(user, code)
	if err != nil {
		return User{}, "", err
	}
	token, session, err := s.session(user.ID, ip, agent, false)
	if err != nil {
		return User{}, "", err
	}
	err = s.Store.ApplyTOTP(ctx, TOTPChange{UserID: user.ID, Secret: user.Secret, Step: step, Action: "login", CurrentHash: TokenHash(pendingToken), NewSession: session, Now: s.Now()})
	return user, token, err
}
