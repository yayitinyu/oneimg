package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"oneimg/backend/config"
	"oneimg/backend/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func TestAuthMiddlewareRejectsLegacyGuestSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testConfig := &config.Config{
		SessionSecret: "oneimg-test-session-secret-32-bytes",
		SessionSecure: false,
	}

	router := gin.New()
	router.Use(SessionMiddleware(testConfig))
	router.POST("/seed-guest", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("logged_in", true)
		session.Set("user_id", 42)
		session.Set("user_role", models.RoleGuest)
		session.Set("username", "legacy-guest")
		session.Set("is_guest", true)
		if err := session.Save(); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})
	router.GET("/protected", AuthMiddleware(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	seedRecorder := httptest.NewRecorder()
	router.ServeHTTP(seedRecorder, httptest.NewRequest(http.MethodPost, "/seed-guest", nil))
	if seedRecorder.Code != http.StatusNoContent || len(seedRecorder.Result().Cookies()) == 0 {
		t.Fatalf("failed to seed guest session: status=%d", seedRecorder.Code)
	}

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.AddCookie(seedRecorder.Result().Cookies()[0])
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("guest session returned %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(recorder.Body.String(), "会话类型已停用") {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}
