package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// TestHelper sets up test environment with required keys
func setupTestKeys(t *testing.T) (teardown func()) {
	// Generate valid 32-byte keys in base64 format
	encKey := make([]byte, 32)
	jwtKey := make([]byte, 32)

	if _, err := rand.Read(encKey); err != nil {
		t.Fatalf("Failed to generate encryption key: %v", err)
	}
	if _, err := rand.Read(jwtKey); err != nil {
		t.Fatalf("Failed to generate JWT key: %v", err)
	}

	encKeyB64 := base64.StdEncoding.EncodeToString(encKey)
	jwtKeyB64 := base64.StdEncoding.EncodeToString(jwtKey)

	// Save originals
	origEncKey := os.Getenv("ENCRYPTION_KEY")
	origJWTSecret := os.Getenv("JWT_SECRET")

	// Set test keys
	os.Setenv("ENCRYPTION_KEY", encKeyB64)
	os.Setenv("JWT_SECRET", jwtKeyB64)

	// Return teardown function
	return func() {
		if origEncKey != "" {
			os.Setenv("ENCRYPTION_KEY", origEncKey)
		} else {
			os.Unsetenv("ENCRYPTION_KEY")
		}
		if origJWTSecret != "" {
			os.Setenv("JWT_SECRET", origJWTSecret)
		} else {
			os.Unsetenv("JWT_SECRET")
		}
	}
}

func TestNewSecureAuth(t *testing.T) {
	teardown := setupTestKeys(t)
	defer teardown()
	auth, err := NewSecureAuth()
	if err != nil {
		t.Fatalf("Failed to create secure auth: %v", err)
	}

	if auth == nil {
		t.Fatal("Expected auth instance, got nil")
	}

	if len(auth.encryptionKey) != 32 {
		t.Errorf("Expected encryption key length 32, got %d", len(auth.encryptionKey))
	}

	if len(auth.signingKey) != 32 {
		t.Errorf("Expected signing key length 32, got %d", len(auth.signingKey))
	}

	if auth.sessionStore == nil {
		t.Fatal("Expected session store to be initialized")
	}
}

func TestCreateEncryptedToken(t *testing.T) {
	teardown := setupTestKeys(t)
	defer teardown()
	auth, err := NewSecureAuth()
	if err != nil {
		t.Fatalf("Failed to create secure auth: %v", err)
	}

	userID := "test-user-123"
	email := "test@example.com"

	accessToken, refreshToken, err := auth.CreateEncryptedToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to create encrypted token: %v", err)
	}

	if accessToken == "" {
		t.Fatal("Expected access token, got empty string")
	}

	if refreshToken == "" {
		t.Fatal("Expected refresh token, got empty string")
	}

	// Verify session was created
	if len(auth.sessionStore) != 1 {
		t.Errorf("Expected 1 session, got %d", len(auth.sessionStore))
	}

	// Find the session
	var session *Session
	for _, s := range auth.sessionStore {
		if s.UserID == userID && s.Email == email {
			session = s
			break
		}
	}

	if session == nil {
		t.Fatal("Expected session to be created")
	}

	if session.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, session.UserID)
	}

	if session.Email != email {
		t.Errorf("Expected email %s, got %s", email, session.Email)
	}

	if time.Until(session.ExpiresAt) < 23*time.Hour {
		t.Error("Expected session to expire in ~24 hours")
	}
}

func TestValidateEncryptedToken(t *testing.T) {
	teardown := setupTestKeys(t)
	defer teardown()
	auth, err := NewSecureAuth()
	if err != nil {
		t.Fatalf("Failed to create secure auth: %v", err)
	}

	userID := "test-user-123"
	email := "test@example.com"

	accessToken, _, err := auth.CreateEncryptedToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to create encrypted token: %v", err)
	}

	// Valid token should work
	claims, err := auth.ValidateEncryptedToken(accessToken)
	if err != nil {
		t.Fatalf("Failed to validate encrypted token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}

	if claims.SessionID == "" {
		t.Fatal("Expected session ID, got empty string")
	}

	// Invalid token should fail
	_, err = auth.ValidateEncryptedToken("invalid-token")
	if err == nil {
		t.Fatal("Expected error for invalid token")
	}

	// Empty token should fail
	_, err = auth.ValidateEncryptedToken("")
	if err == nil {
		t.Fatal("Expected error for empty token")
	}
}

func TestRefreshToken(t *testing.T) {
	teardown := setupTestKeys(t)
	defer teardown()
	auth, err := NewSecureAuth()
	if err != nil {
		t.Fatalf("Failed to create secure auth: %v", err)
	}

	userID := "test-user-123"
	email := "test@example.com"

	_, refreshToken, err := auth.CreateEncryptedToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to create encrypted token: %v", err)
	}

	// Valid refresh token should work
	newAccessToken, err := auth.RefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}

	if newAccessToken == "" {
		t.Fatal("Expected new access token, got empty string")
	}

	// Invalid refresh token should fail
	_, err = auth.RefreshToken("invalid-refresh-token")
	if err == nil {
		t.Fatal("Expected error for invalid refresh token")
	}

	// Empty refresh token should fail
	_, err = auth.RefreshToken("")
	if err == nil {
		t.Fatal("Expected error for empty refresh token")
	}
}

func TestRevokeSession(t *testing.T) {
	teardown := setupTestKeys(t)
	defer teardown()
	auth, err := NewSecureAuth()
	if err != nil {
		t.Fatalf("Failed to create secure auth: %v", err)
	}

	userID := "test-user-123"
	email := "test@example.com"

	accessToken, _, err := auth.CreateEncryptedToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to create encrypted token: %v", err)
	}

	// Validate token works
	claims, err := auth.ValidateEncryptedToken(accessToken)
	if err != nil {
		t.Fatalf("Failed to validate encrypted token: %v", err)
	}

	sessionID := claims.SessionID

	// Revoke session
	auth.RevokeSession(sessionID)

	// Token should now be invalid
	_, err = auth.ValidateEncryptedToken(accessToken)
	if err == nil {
		t.Fatal("Expected error for revoked session")
	}
}

func TestRevokeAllUserSessions(t *testing.T) {
	teardown := setupTestKeys(t)
	defer teardown()
	auth, err := NewSecureAuth()
	if err != nil {
		t.Fatalf("Failed to create secure auth: %v", err)
	}

	userID := "test-user-123"
	email := "test@example.com"

	// Create multiple sessions for the same user
	accessToken1, _, err := auth.CreateEncryptedToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to create first encrypted token: %v", err)
	}

	accessToken2, _, err := auth.CreateEncryptedToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to create second encrypted token: %v", err)
	}

	// Both tokens should work
	_, err = auth.ValidateEncryptedToken(accessToken1)
	if err != nil {
		t.Fatalf("Failed to validate first token: %v", err)
	}

	_, err = auth.ValidateEncryptedToken(accessToken2)
	if err != nil {
		t.Fatalf("Failed to validate second token: %v", err)
	}

	// Revoke all sessions for user
	auth.RevokeAllUserSessions(userID)

	// Both tokens should now be invalid
	_, err = auth.ValidateEncryptedToken(accessToken1)
	if err == nil {
		t.Fatal("Expected error for first revoked session")
	}

	_, err = auth.ValidateEncryptedToken(accessToken2)
	if err == nil {
		t.Fatal("Expected error for second revoked session")
	}
}

func TestTokenExpiration(t *testing.T) {
	teardown := setupTestKeys(t)
	defer teardown()
	auth, err := NewSecureAuth()
	if err != nil {
		t.Fatalf("Failed to create secure auth: %v", err)
	}

	userID := "test-user-123"
	email := "test@example.com"

	accessToken, _, err := auth.CreateEncryptedToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to create encrypted token: %v", err)
	}

	// Find the session and manually expire it
	var sessionID string
	for id, session := range auth.sessionStore {
		if session.UserID == userID {
			sessionID = id
			session.ExpiresAt = time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
			break
		}
	}

	if sessionID == "" {
		t.Fatal("Expected to find session")
	}

	// Token should now be invalid due to expiration
	_, err = auth.ValidateEncryptedToken(accessToken)
	if err == nil {
		t.Fatal("Expected error for expired session")
	}

	// Session should be removed from store
	if _, exists := auth.sessionStore[sessionID]; exists {
		t.Fatal("Expected expired session to be removed from store")
	}
}

func TestCSRFTokenValidation(t *testing.T) {
	sessionID := "test-session-123"

	// Generate CSRF token
	csrfToken := GenerateCSRFToken(sessionID)
	if csrfToken == "" {
		t.Fatal("Expected CSRF token, got empty string")
	}

	// Valid CSRF token should work
	if !ValidateCSRFToken(csrfToken, sessionID) {
		t.Fatal("Expected valid CSRF token to pass validation")
	}

	// Invalid CSRF token should fail
	if ValidateCSRFToken("invalid-csrf-token", sessionID) {
		t.Fatal("Expected invalid CSRF token to fail validation")
	}

	// Wrong session ID should fail
	if ValidateCSRFToken(csrfToken, "wrong-session-id") {
		t.Fatal("Expected CSRF token with wrong session ID to fail validation")
	}

	// Empty CSRF token should fail
	if ValidateCSRFToken("", sessionID) {
		t.Fatal("Expected empty CSRF token to fail validation")
	}

	// Empty session ID should fail
	if ValidateCSRFToken(csrfToken, "") {
		t.Fatal("Expected CSRF token with empty session ID to fail validation")
	}
}

func TestCreateSecureCookie(t *testing.T) {
	name := "test-cookie"
	value := "test-value"
	maxAge := 3600

	cookie := CreateSecureCookie(name, value, maxAge)

	expectedFields := []string{"Name", "Value", "Path", "MaxAge", "HttpOnly", "Secure", "SameSite"}
	for _, field := range expectedFields {
		if _, exists := cookie[field]; !exists {
			t.Errorf("Expected cookie to have field %s", field)
		}
	}

	if cookie["Name"] != name {
		t.Errorf("Expected cookie name %s, got %s", name, cookie["Name"])
	}

	if cookie["Value"] != value {
		t.Errorf("Expected cookie value %s, got %s", value, cookie["Value"])
	}

	if cookie["Path"] != "/" {
		t.Errorf("Expected cookie path /, got %s", cookie["Path"])
	}

	if cookie["HttpOnly"] != "true" {
		t.Errorf("Expected HttpOnly true, got %s", cookie["HttpOnly"])
	}

	if cookie["Secure"] != "true" {
		t.Errorf("Expected Secure true, got %s", cookie["Secure"])
	}

	if cookie["SameSite"] != "Strict" {
		t.Errorf("Expected SameSite Strict, got %s", cookie["SameSite"])
	}
}

func TestEncryptionDecryption(t *testing.T) {
	teardown := setupTestKeys(t)
	defer teardown()
	auth, err := NewSecureAuth()
	if err != nil {
		t.Fatalf("Failed to create secure auth: %v", err)
	}

	plaintext := []byte("This is a test message for encryption")

	// Encrypt
	ciphertext, err := auth.encrypt(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	if len(ciphertext) == 0 {
		t.Fatal("Expected ciphertext, got empty")
	}

	// Decrypt
	decrypted, err := auth.decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Expected decrypted text %s, got %s", string(plaintext), string(decrypted))
	}

	// Invalid ciphertext should fail
	_, err = auth.decrypt([]byte("invalid-ciphertext"))
	if err == nil {
		t.Fatal("Expected error for invalid ciphertext")
	}

	// Empty ciphertext should fail
	_, err = auth.decrypt([]byte{})
	if err == nil {
		t.Fatal("Expected error for empty ciphertext")
	}
}

func TestConcurrentAccess(t *testing.T) {
	teardown := setupTestKeys(t)
	defer teardown()
	auth, err := NewSecureAuth()
	if err != nil {
		t.Fatalf("Failed to create secure auth: %v", err)
	}

	// Test concurrent token creation and validation
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(userNum int) {
			defer func() { done <- true }()

			userID := fmt.Sprintf("user-%d", userNum)
			email := fmt.Sprintf("user%d@example.com", userNum)

			// Create token
			accessToken, _, err := auth.CreateEncryptedToken(userID, email)
			if err != nil {
				t.Errorf("Failed to create token for user %d: %v", userNum, err)
				return
			}

			// Validate token
			claims, err := auth.ValidateEncryptedToken(accessToken)
			if err != nil {
				t.Errorf("Failed to validate token for user %d: %v", userNum, err)
				return
			}

			if claims.UserID != userID {
				t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all sessions were created
	if len(auth.sessionStore) != 10 {
		t.Errorf("Expected 10 sessions, got %d", len(auth.sessionStore))
	}
}

// TestGetOrGenerateKeySecurityFix tests that keys are not printed to stdout
// This test verifies the security fix where keys are no longer auto-generated and printed
func TestGetOrGenerateKeyMissingEnvironment(t *testing.T) {
	// Save original env vars
	origEncKey := os.Getenv("ENCRYPTION_KEY")
	origJWTSecret := os.Getenv("JWT_SECRET")
	defer func() {
		// Restore original env vars
		if origEncKey != "" {
			os.Setenv("ENCRYPTION_KEY", origEncKey)
		} else {
			os.Unsetenv("ENCRYPTION_KEY")
		}
		if origJWTSecret != "" {
			os.Setenv("JWT_SECRET", origJWTSecret)
		} else {
			os.Unsetenv("JWT_SECRET")
		}
	}()

	// Clear environment variables to test missing key behavior
	os.Unsetenv("ENCRYPTION_KEY")
	os.Unsetenv("JWT_SECRET")

	// Creating SecureAuth without environment variables should fail
	auth, err := NewSecureAuth()
	if err == nil {
		t.Fatal("Expected error when ENCRYPTION_KEY and JWT_SECRET are not set")
	}

	if auth != nil {
		t.Fatal("Expected auth to be nil when initialization fails")
	}

	// Verify error message mentions environment variable
	if !strings.Contains(err.Error(), "ENCRYPTION_KEY") && !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Errorf("Expected error to mention missing environment variable, got: %v", err)
	}
}

// TestValidKeyFormat tests that keys must be proper base64-encoded 32-byte values
func TestGetOrGenerateKeyValidFormat(t *testing.T) {
	// Save original env vars
	origEncKey := os.Getenv("ENCRYPTION_KEY")
	origJWTSecret := os.Getenv("JWT_SECRET")
	defer func() {
		if origEncKey != "" {
			os.Setenv("ENCRYPTION_KEY", origEncKey)
		} else {
			os.Unsetenv("ENCRYPTION_KEY")
		}
		if origJWTSecret != "" {
			os.Setenv("JWT_SECRET", origJWTSecret)
		} else {
			os.Unsetenv("JWT_SECRET")
		}
	}()

	// Set invalid base64
	os.Setenv("ENCRYPTION_KEY", "not-valid-base64!!!")
	os.Setenv("JWT_SECRET", "dGVzdA==") // Valid base64 but wrong length

	auth, err := NewSecureAuth()
	if err == nil {
		t.Fatal("Expected error with invalid key format")
	}

	if auth != nil {
		t.Fatal("Expected auth to be nil")
	}

	if !strings.Contains(err.Error(), "invalid key") {
		t.Errorf("Expected 'invalid key' in error message, got: %v", err)
	}
}
