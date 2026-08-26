package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/kodehat/portkey/internal/config"
)

// WA is the shared WebAuthn instance.
var WA *webauthn.WebAuthn

// User implements webauthn.User for a single owner.
type User struct {
	ID          []byte               `json:"id"`
	Name        string               `json:"name"`
	DisplayName string               `json:"displayName"`
	Credentials []webauthn.Credential `json:"credentials"`
}

func (u *User) WebAuthnID() []byte                        { return u.ID }
func (u *User) WebAuthnName() string                       { return u.Name }
func (u *User) WebAuthnDisplayName() string                { return u.DisplayName }
func (u *User) WebAuthnCredentials() []webauthn.Credential { return u.Credentials }

type persisted struct {
	User          *User  `json:"user,omitempty"`
	SessionSecret string `json:"sessionSecret"`
}

var (
	mu          sync.RWMutex
	state       persisted
	regSession  *webauthn.SessionData
	authSession *webauthn.SessionData
	pendingUser *User
)

// Init sets up the WebAuthn instance and loads stored credentials.
func Init() error {
	var err error
	WA, err = webauthn.New(&webauthn.Config{
		RPDisplayName: "AM-links",
		RPID:          config.C.Auth.RPId,
		RPOrigins:     []string{config.C.Auth.RPOrigin},
	})
	if err != nil {
		return fmt.Errorf("webauthn init: %w", err)
	}
	return load()
}

func load() error {
	mu.Lock()
	defer mu.Unlock()
	raw, err := os.ReadFile(config.C.Auth.CredentialsFile)
	if os.IsNotExist(err) || len(raw) == 0 {
		return initStore()
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, &state)
}

func initStore() error {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return err
	}
	state = persisted{SessionSecret: hex.EncodeToString(b)}
	return save()
}

func save() error {
	b, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(config.C.Auth.CredentialsFile, b, 0600)
}

// HasCredentials returns true if a passkey is already registered.
func HasCredentials() bool {
	mu.RLock()
	defer mu.RUnlock()
	return state.User != nil && len(state.User.Credentials) > 0
}

func getUser() *User {
	mu.RLock()
	defer mu.RUnlock()
	return state.User
}

func updateCredential(cred webauthn.Credential) error {
	mu.Lock()
	defer mu.Unlock()
	for i, c := range state.User.Credentials {
		if string(c.ID) == string(cred.ID) {
			state.User.Credentials[i] = cred
			return save()
		}
	}
	return nil
}

// BeginRegistration starts a passkey registration ceremony.
// Returns an error if a credential is already registered.
func BeginRegistration() (*protocol.CredentialCreation, error) {
	if HasCredentials() {
		return nil, fmt.Errorf("already registered")
	}
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return nil, err
	}
	user := &User{
		ID:          id,
		Name:        "owner",
		DisplayName: "Owner",
		Credentials: []webauthn.Credential{},
	}
	options, session, err := WA.BeginRegistration(
		user,
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			AuthenticatorAttachment: protocol.Platform,
			ResidentKey:             protocol.ResidentKeyRequirementRequired,
			UserVerification:        protocol.VerificationRequired,
		}),
		webauthn.WithConveyancePreference(protocol.PreferNoAttestation),
	)
	if err != nil {
		return nil, err
	}
	mu.Lock()
	regSession = session
	pendingUser = user
	mu.Unlock()
	return options, nil
}

// FinishRegistration completes the ceremony and persists the new credential.
func FinishRegistration(r *http.Request) error {
	mu.RLock()
	session := regSession
	user := pendingUser
	mu.RUnlock()
	if session == nil || user == nil {
		return fmt.Errorf("no active registration session")
	}
	cred, err := WA.FinishRegistration(user, *session, r)
	if err != nil {
		return err
	}
	user.Credentials = []webauthn.Credential{*cred}
	mu.Lock()
	state.User = user
	regSession = nil
	pendingUser = nil
	err = save()
	mu.Unlock()
	return err
}

// BeginLogin starts a discoverable-credential login ceremony (no user lookup needed).
func BeginLogin() (*protocol.CredentialAssertion, error) {
	if !HasCredentials() {
		return nil, fmt.Errorf("no credentials registered")
	}
	options, session, err := WA.BeginDiscoverableLogin(
		webauthn.WithUserVerification(protocol.VerificationRequired),
	)
	if err != nil {
		return nil, err
	}
	mu.Lock()
	authSession = session
	mu.Unlock()
	return options, nil
}

// FinishLogin validates the assertion and updates the credential counter.
func FinishLogin(r *http.Request) error {
	mu.RLock()
	session := authSession
	mu.RUnlock()
	if session == nil {
		return fmt.Errorf("no active login session")
	}
	cred, err := WA.FinishDiscoverableLogin(
		func(rawID, userHandle []byte) (webauthn.User, error) {
			u := getUser()
			if u == nil {
				return nil, fmt.Errorf("no registered user")
			}
			return u, nil
		},
		*session,
		r,
	)
	if err != nil {
		return err
	}
	mu.Lock()
	authSession = nil
	mu.Unlock()
	return updateCredential(*cred)
}

// --- HMAC-signed session tokens ---

// NewSessionToken returns a token valid for 30 days.
func NewSessionToken() string {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	return ts + "." + computeMAC(ts)
}

// ValidSession checks authenticity and expiry.
func ValidSession(token string) bool {
	i := strings.LastIndexByte(token, '.')
	if i < 0 {
		return false
	}
	ts, sig := token[:i], token[i+1:]
	if !hmac.Equal([]byte(computeMAC(ts)), []byte(sig)) {
		return false
	}
	t, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return false
	}
	return time.Since(time.Unix(t, 0)) < 30*24*time.Hour
}

func computeMAC(msg string) string {
	mu.RLock()
	secret := state.SessionSecret
	mu.RUnlock()
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))
}
