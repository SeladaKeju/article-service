package route

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SeladaKeju/article-service/api/controller"
)

func TestSwaggerDirectoryRedirectsToIndex(t *testing.T) {
	for _, test := range []struct {
		path, location string
		status         int
	}{
		{"/swagger", "/swagger/", http.StatusMovedPermanently},
		{"/swagger/", "/swagger/index.html", http.StatusFound},
	} {
		w := httptest.NewRecorder()
		New(controller.ArticleController{}, time.Second).ServeHTTP(w, httptest.NewRequest(http.MethodGet, test.path, nil))
		if w.Code != test.status || w.Header().Get("Location") != test.location {
			t.Fatalf("path=%q status=%d location=%q", test.path, w.Code, w.Header().Get("Location"))
		}
	}
}
