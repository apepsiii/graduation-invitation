package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	SessionCookieName = "admin_session"
	SessionMaxAge     = 24 * time.Hour
)

type SessionManager struct {
	secretKey string
}

func NewSessionManager(secretKey string) *SessionManager {
	if secretKey == "" {
		secretKey = "default-secret-key-change-in-production"
	}
	return &SessionManager{secretKey: secretKey}
}

func (sm *SessionManager) createToken(username string, expiry time.Time) string {
	expStr := strconv.FormatInt(expiry.Unix(), 10)
	payload := username + ":" + expStr
	mac := hmac.New(sha256.New, []byte(sm.secretKey))
	mac.Write([]byte(payload))
	sig := base64.URLEncoding.EncodeToString(mac.Sum(nil))
	return base64.URLEncoding.EncodeToString([]byte(username)) + "." + expStr + "." + sig
}

func (sm *SessionManager) VerifyToken(token string) (string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", false
	}

	usernameBytes, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", false
	}
	username := string(usernameBytes)
	expStr := parts[1]
	payload := username + ":" + expStr

	mac := hmac.New(sha256.New, []byte(sm.secretKey))
	mac.Write([]byte(payload))
	expectedSig := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return "", false
	}

	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return "", false
	}

	return username, true
}

func (sm *SessionManager) SetSession(c *gin.Context, username string) {
	expiry := time.Now().Add(SessionMaxAge)
	token := sm.createToken(username, expiry)
	c.SetCookie(SessionCookieName, token, int(SessionMaxAge.Seconds()), "/", "", false, true)
}

func (sm *SessionManager) ClearSession(c *gin.Context) {
	c.SetCookie(SessionCookieName, "", -1, "/", "", false, true)
}

func (sm *SessionManager) GetSessionUser(c *gin.Context) (string, bool) {
	token, err := c.Cookie(SessionCookieName)
	if err != nil || token == "" {
		return "", false
	}
	return sm.VerifyToken(token)
}

func (sm *SessionManager) SessionAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, ok := sm.GetSessionUser(c)
		if !ok {
			accept := c.GetHeader("Accept")
			if strings.Contains(accept, "text/html") || c.Request.Method == "GET" && !strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.Redirect(http.StatusFound, "/admin/login")
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			}
			c.Abort()
			return
		}
		c.Set("username", username)
		c.Next()
	}
}

func (sm *SessionManager) APIAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, ok := sm.GetSessionUser(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Set("username", username)
		c.Next()
	}
}
