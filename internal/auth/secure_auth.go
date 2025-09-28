package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SecureAuth handles encrypted JWT tokens and session management
type SecureAuth struct {
	encryptionKey []byte
	signingKey    []byte
	sessionStore  map[string]*Session
	mutex         sync.RWMutex
}

// Session represents a secure user session
type Session struct {
	UserID       string    `json:"user_id"`
	Email        string    `json:"email"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	LastUsed     time.Time `json:"last_used"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
}

// TokenClaims represents encrypted JWT claims
type TokenClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

// NewSecureAuth creates a new secure auth instance
func NewSecureAuth() (*SecureAuth, error) {
	// Generate or load encryption key
	encryptionKey, err := getOrGenerateKey("ENCRYPTION_KEY", 32)
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}

	// Generate or load signing key
	signingKey, err := getOrGenerateKey("JWT_SECRET", 32)
	if err != nil {
		return nil, fmt.Errorf("failed to get signing key: %w", err)
	}

	return &SecureAuth{
		encryptionKey: encryptionKey,
		signingKey:    signingKey,
		sessionStore:  make(map[string]*Session),
	}, nil
}

// getOrGenerateKey gets a key from environment or generates a secure one
func getOrGenerateKey(envVar string, size int) ([]byte, error) {
	keyStr := os.Getenv(envVar)
	if keyStr == "" {
		// Generate a secure random key
		key := make([]byte, size)
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("failed to generate random key: %w", err)
		}
		// Convert to base64 for storage
		keyStr = base64.StdEncoding.EncodeToString(key)
		fmt.Printf("⚠️  Generated new %s: %s\n", envVar, keyStr)
		fmt.Printf("   Set this in your environment: export %s=%s\n", envVar, keyStr)
		return key, nil
	}

	// Decode base64 key
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid key format: %w", err)
	}

	if len(key) != size {
		return nil, fmt.Errorf("key must be %d bytes", size)
	}

	return key, nil
}

// encrypt encrypts data using AES-GCM
func (sa *SecureAuth) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(sa.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-GCM
func (sa *SecureAuth) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(sa.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// CreateEncryptedToken creates an encrypted JWT token
func (sa *SecureAuth) CreateEncryptedToken(userID, email string) (string, string, error) {
	// Create session
	sessionID := generateSessionID()
	refreshToken := generateRefreshToken()

	session := &Session{
		UserID:       userID,
		Email:        email,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour), // 24 hours
		CreatedAt:    time.Now(),
		LastUsed:     time.Now(),
	}

	// Store session
	sa.mutex.Lock()
	sa.sessionStore[sessionID] = session
	sa.mutex.Unlock()

	// Create JWT claims
	claims := TokenClaims{
		UserID:    userID,
		Email:     email,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // 15 minutes
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "bffgen",
		},
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(sa.signingKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign token: %w", err)
	}

	// Encrypt the token
	encryptedToken, err := sa.encrypt([]byte(tokenString))
	if err != nil {
		return "", "", fmt.Errorf("failed to encrypt token: %w", err)
	}

	// Encode to base64 for transport
	encodedToken := base64.StdEncoding.EncodeToString(encryptedToken)

	return encodedToken, refreshToken, nil
}

// ValidateEncryptedToken validates and decrypts a JWT token
func (sa *SecureAuth) ValidateEncryptedToken(encryptedToken string) (*TokenClaims, error) {
	// Decode base64
	tokenBytes, err := base64.StdEncoding.DecodeString(encryptedToken)
	if err != nil {
		return nil, fmt.Errorf("invalid token format: %w", err)
	}

	// Decrypt token
	decryptedBytes, err := sa.decrypt(tokenBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	// Parse JWT token
	token, err := jwt.ParseWithClaims(string(decryptedBytes), &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return sa.signingKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate session
	sa.mutex.RLock()
	session, exists := sa.sessionStore[claims.SessionID]
	sa.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		sa.mutex.Lock()
		delete(sa.sessionStore, claims.SessionID)
		sa.mutex.Unlock()
		return nil, fmt.Errorf("session expired")
	}

	// Update last used
	sa.mutex.Lock()
	session.LastUsed = time.Now()
	sa.mutex.Unlock()

	return claims, nil
}

// RefreshToken refreshes an access token using refresh token
func (sa *SecureAuth) RefreshToken(refreshToken string) (string, error) {
	// Find session by refresh token
	var session *Session
	var sessionID string

	sa.mutex.RLock()
	for id, s := range sa.sessionStore {
		if s.RefreshToken == refreshToken {
			session = s
			sessionID = id
			break
		}
	}
	sa.mutex.RUnlock()

	if session == nil {
		return "", fmt.Errorf("invalid refresh token")
	}

	if time.Now().After(session.ExpiresAt) {
		sa.mutex.Lock()
		delete(sa.sessionStore, sessionID)
		sa.mutex.Unlock()
		return "", fmt.Errorf("refresh token expired")
	}

	// Create new access token
	newToken, _, err := sa.CreateEncryptedToken(session.UserID, session.Email)
	if err != nil {
		return "", fmt.Errorf("failed to create new token: %w", err)
	}

	return newToken, nil
}

// RevokeSession revokes a user session
func (sa *SecureAuth) RevokeSession(sessionID string) {
	sa.mutex.Lock()
	delete(sa.sessionStore, sessionID)
	sa.mutex.Unlock()
}

// RevokeAllUserSessions revokes all sessions for a user
func (sa *SecureAuth) RevokeAllUserSessions(userID string) {
	sa.mutex.Lock()
	defer sa.mutex.Unlock()

	for sessionID, session := range sa.sessionStore {
		if session.UserID == userID {
			delete(sa.sessionStore, sessionID)
		}
	}
}

// generateSessionID generates a secure session ID
func generateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// generateRefreshToken generates a secure refresh token
func generateRefreshToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// CreateSecureCookie creates a secure HTTP cookie
func CreateSecureCookie(name, value string, maxAge int) map[string]string {
	return map[string]string{
		"Name":     name,
		"Value":    value,
		"Path":     "/",
		"MaxAge":   fmt.Sprintf("%d", maxAge),
		"HttpOnly": "true",
		"Secure":   "true",
		"SameSite": "Strict",
	}
}

// ValidateCSRFToken validates CSRF token
func ValidateCSRFToken(token, sessionID string) bool {
	// Generate expected CSRF token from session ID
	hash := sha256.Sum256([]byte(sessionID + "csrf"))
	expectedToken := base64.StdEncoding.EncodeToString(hash[:])
	return token == expectedToken
}

// GenerateCSRFToken generates CSRF token for session
func GenerateCSRFToken(sessionID string) string {
	hash := sha256.Sum256([]byte(sessionID + "csrf"))
	return base64.StdEncoding.EncodeToString(hash[:])
}
