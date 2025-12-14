// Package values provides domain value objects with validation.
package values

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/LarsArtmann/template-arch-lint/pkg/errors"
)

// SessionToken represents a user session token.
type SessionToken struct {
	value string
	expires time.Time
}

// NewSessionToken creates a new session token with expiration.
func NewSessionToken(duration time.Duration) (SessionToken, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return SessionToken{}, errors.NewInfrastructureError("session_token", "generate", err)
	}
	
	token := fmt.Sprintf("%x", bytes)
	expires := time.Now().Add(duration)
	
	return SessionToken{
		value:   token,
		expires: expires,
	}, nil
}

// NewSessionTokenFromValue creates a session token from existing value.
func NewSessionTokenFromValue(value string, expires time.Time) (SessionToken, error) {
	if err := validateSessionToken(value); err != nil {
		return SessionToken{}, err
	}
	
	return SessionToken{
		value:   value,
		expires: expires,
	}, nil
}

// validateSessionToken validates session token format.
func validateSessionToken(token string) error {
	if len(token) < 32 {
		return errors.NewDomainValidationError("session_token", "token too short (minimum 32 characters)")
	}
	
	if len(token) > 256 {
		return errors.NewDomainValidationError("session_token", "token too long (maximum 256 characters)")
	}
	
	// Should be hexadecimal characters only
	matched, err := regexp.MatchString(`^[a-fA-F0-9]+$`, token)
	if err != nil {
		return errors.NewInfrastructureError("session_token", "validate", err)
	}
	
	if !matched {
		return errors.NewDomainValidationError("session_token", "token must contain only hexadecimal characters")
	}
	
	return nil
}

// String returns the string representation of session token.
func (t SessionToken) String() string {
	return t.value
}

// Expires returns the expiration time of session token.
func (t SessionToken) Expires() time.Time {
	return t.expires
}

// IsExpired checks if the session token has expired.
func (t SessionToken) IsExpired() bool {
	return time.Now().After(t.expires)
}

// IsValid checks if the session token is still valid.
func (t SessionToken) IsValid() bool {
	return !t.IsExpired() && t.value != ""
}

// MarshalJSON implements json.Marshaler interface.
func (t SessionToken) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Token   string    `json:"token"`
		Expires time.Time `json:"expires"`
	}{
		Token:   t.value,
		Expires: t.expires,
	})
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (t *SessionToken) UnmarshalJSON(data []byte) error {
	var session struct {
		Token   string    `json:"token"`
		Expires time.Time `json:"expires"`
	}
	
	if err := json.Unmarshal(data, &session); err != nil {
		return err
	}
	
	token, err := NewSessionTokenFromValue(session.Token, session.Expires)
	if err != nil {
		return err
	}
	
	*t = token
	return nil
}

// AuditTrail represents an audit trail entry.
type AuditTrail struct {
	userID    string
	action    string
	resource  string
	timestamp time.Time
	ip        string
	userAgent string
	metadata  map[string]string
}

// NewAuditTrail creates a new audit trail entry.
func NewAuditTrail(userID, action, resource, ip, userAgent string) AuditTrail {
	return AuditTrail{
		userID:    userID,
		action:    action,
		resource:  resource,
		timestamp: time.Now().UTC(),
		ip:        ip,
		userAgent: userAgent,
		metadata:  make(map[string]string),
	}
}

// UserID returns the user ID from audit trail.
func (a AuditTrail) UserID() string {
	return a.userID
}

// Action returns the action from audit trail.
func (a AuditTrail) Action() string {
	return a.action
}

// Resource returns the resource from audit trail.
func (a AuditTrail) Resource() string {
	return a.resource
}

// Timestamp returns the timestamp from audit trail.
func (a AuditTrail) Timestamp() time.Time {
	return a.timestamp
}

// IP returns the IP address from audit trail.
func (a AuditTrail) IP() string {
	return a.ip
}

// UserAgent returns the user agent from audit trail.
func (a AuditTrail) UserAgent() string {
	return a.userAgent
}

// Metadata returns the metadata from audit trail.
func (a AuditTrail) Metadata() map[string]string {
	return a.metadata
}

// AddMetadata adds metadata to audit trail.
func (a *AuditTrail) AddMetadata(key, value string) {
	if a.metadata == nil {
		a.metadata = make(map[string]string)
	}
	a.metadata[key] = value
}

// MarshalJSON implements json.Marshaler interface.
func (a AuditTrail) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		UserID    string            `json:"user_id"`
		Action    string            `json:"action"`
		Resource  string            `json:"resource"`
		Timestamp time.Time         `json:"timestamp"`
		IP        string            `json:"ip"`
		UserAgent string            `json:"user_agent"`
		Metadata  map[string]string `json:"metadata"`
	}{
		UserID:    a.userID,
		Action:    a.action,
		Resource:  a.resource,
		Timestamp: a.timestamp,
		IP:        a.ip,
		UserAgent: a.userAgent,
		Metadata:  a.metadata,
	})
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (a *AuditTrail) UnmarshalJSON(data []byte) error {
	var audit struct {
		UserID    string            `json:"user_id"`
		Action    string            `json:"action"`
		Resource  string            `json:"resource"`
		Timestamp time.Time         `json:"timestamp"`
		IP        string            `json:"ip"`
		UserAgent string            `json:"user_agent"`
		Metadata  map[string]string `json:"metadata"`
	}
	
	if err := json.Unmarshal(data, &audit); err != nil {
		return err
	}
	
	a.userID = audit.UserID
	a.action = audit.Action
	a.resource = audit.Resource
	a.timestamp = audit.Timestamp
	a.ip = audit.IP
	a.userAgent = audit.UserAgent
	a.metadata = audit.Metadata
	
	return nil
}