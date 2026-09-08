package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestTimeoutAddsRequestDeadline(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Timeout(time.Second))
	router.GET("/", func(c *gin.Context) {
		if _, ok := c.Request.Context().Deadline(); !ok {
			t.Fatal("request context has no deadline")
		}
		c.Status(http.StatusNoContent)
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
}
