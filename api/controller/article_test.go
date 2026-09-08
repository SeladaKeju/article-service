package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SeladaKeju/article-service.git/domain"
	"github.com/SeladaKeju/article-service.git/repository"
	"github.com/SeladaKeju/article-service.git/usecase"
	"github.com/gin-gonic/gin"
)

type articleStoreStub struct {
	article domain.Article
	err     error
}

func (s *articleStoreStub) Create(_ context.Context, article domain.Article) (domain.Article, error) {
	s.article = article
	if s.err != nil {
		return domain.Article{}, s.err
	}
	return article, nil
}

func TestCreateArticle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &articleStoreStub{}
	router := gin.New()
	router.POST("/articles", NewArticleController(usecase.NewArticleUsecase(store)).Create)

	request := httptest.NewRequest(http.MethodPost, "/articles", strings.NewReader(`{"author_id":" 550E8400-E29B-41D4-A716-446655440000 ","title":" Go ","body":" Body "}`))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if store.article.AuthorID != "550e8400-e29b-41d4-a716-446655440000" || store.article.Title != " Go " || store.article.Body != " Body " {
		t.Fatalf("stored article = %#v", store.article)
	}

	var body struct {
		Data articleResponse `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if _, valid := domain.NormalizeUUIDv4(body.Data.ID); !valid {
		t.Fatalf("id = %q", body.Data.ID)
	}
	if _, err := time.Parse("2006-01-02T15:04:05.000000Z", body.Data.CreatedAt); err != nil {
		t.Fatalf("created_at = %q: %v", body.Data.CreatedAt, err)
	}
}

func TestCreateArticleRejectsInvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &articleStoreStub{}
	router := gin.New()
	router.POST("/articles", NewArticleController(usecase.NewArticleUsecase(store)).Create)

	request := httptest.NewRequest(http.MethodPost, "/articles", strings.NewReader(`{"author_id":"550e8400-e29b-41d4-a716-446655440000","title":"   ","body":"body","extra":true}`))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest || store.article.ID != "" || !strings.Contains(response.Body.String(), `"invalid_request"`) {
		t.Fatalf("status = %d, stored = %#v, body = %s", response.Code, store.article, response.Body.String())
	}
}

func TestCreateArticleReportsMissingAuthor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &articleStoreStub{err: repository.ErrAuthorNotFound}
	router := gin.New()
	router.POST("/articles", NewArticleController(usecase.NewArticleUsecase(store)).Create)

	request := httptest.NewRequest(http.MethodPost, "/articles", strings.NewReader(`{"author_id":"550e8400-e29b-41d4-a716-446655440000","title":"title","body":"body"}`))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"author_not_found"`) {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
