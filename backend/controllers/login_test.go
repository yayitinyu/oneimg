package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"oneimg/backend/config"
	"oneimg/backend/middlewares"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func TestLogoutExpiresSessionCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	originalConfig := config.App
	t.Cleanup(func() { config.App = originalConfig })

	testConfig := &config.Config{
		SessionSecret: "oneimg-test-session-secret-32-bytes",
		SessionSecure: false,
	}
	config.App = testConfig

	router := gin.New()
	router.Use(middlewares.SessionMiddleware(testConfig))
	router.POST("/seed-session", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("logged_in", true)
		if err := session.Save(); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})
	router.POST("/logout", Logout)

	seedRecorder := httptest.NewRecorder()
	router.ServeHTTP(seedRecorder, httptest.NewRequest(http.MethodPost, "/seed-session", nil))
	if seedRecorder.Code != http.StatusNoContent {
		t.Fatalf("seeding session returned %d", seedRecorder.Code)
	}
	seedCookies := seedRecorder.Result().Cookies()
	if len(seedCookies) == 0 {
		t.Fatal("seeding session did not set a cookie")
	}

	logoutRequest := httptest.NewRequest(http.MethodPost, "/logout", nil)
	logoutRequest.AddCookie(seedCookies[0])
	logoutRecorder := httptest.NewRecorder()
	router.ServeHTTP(logoutRecorder, logoutRequest)

	if logoutRecorder.Code != http.StatusOK {
		t.Fatalf("logout returned %d, want %d", logoutRecorder.Code, http.StatusOK)
	}
	setCookie := logoutRecorder.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "oneimg-session=") || !strings.Contains(setCookie, "Max-Age=0") {
		t.Fatalf("logout did not expire the session cookie: %q", setCookie)
	}
}
