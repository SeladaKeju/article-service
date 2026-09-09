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
	"github.com/SeladaKeju/article-service.git/usecase"
	"github.com/gin-gonic/gin"
)

type articleStoreStub struct {
	article    domain.Article
	err        error
	listReturn []domain.Article
}

func (s *articleStoreStub) Create(_ context.Context, article domain.Article) (domain.Article, error) {
	s.article = article
	if s.err != nil {
		return domain.Article{}, s.err
	}
	return article, nil
}

func (s *articleStoreStub) List(_ context.Context, _ domain.ListArticlesParams) ([]domain.Article, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.listReturn, nil
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

	var body ArticleResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if _, valid := domain.NormalizeUUIDv4(body.ID); !valid {
		t.Fatalf("id = %q", body.ID)
	}
	if _, err := time.Parse("2006-01-02T15:04:05.000000Z", body.CreatedAt); err != nil {
		t.Fatalf("created_at = %q: %v", body.CreatedAt, err)
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

func TestCreateArticleRejectsMalformedInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []string{
		`{`,
		`null`,
		`[]`,
		`{"author_id":null,"title":"title","body":"body"}`,
		`{"author_id":123,"title":"title","body":"body"}`,
		`{"author_id":"not-a-uuid","title":"title","body":"body"}`,
		`{"author_id":"550e8400-e29b-41d4-a716-446655440000","title":"title"}`,
	}

	for _, body := range cases {
		store := &articleStoreStub{}
		router := gin.New()
		router.POST("/articles", NewArticleController(usecase.NewArticleUsecase(store)).Create)

		request := httptest.NewRequest(http.MethodPost, "/articles", strings.NewReader(body))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		if response.Code != http.StatusBadRequest || store.article.ID != "" || !strings.Contains(response.Body.String(), `"invalid_request"`) {
			t.Fatalf("body = %q, status = %d, stored = %#v, response = %s", body, response.Code, store.article, response.Body.String())
		}
	}
}

func TestCreateArticleReportsMissingAuthor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &articleStoreStub{err: domain.ErrAuthorNotFound}
	router := gin.New()
	router.POST("/articles", NewArticleController(usecase.NewArticleUsecase(store)).Create)

	request := httptest.NewRequest(http.MethodPost, "/articles", strings.NewReader(`{"author_id":"550e8400-e29b-41d4-a716-446655440000","title":"title","body":"body"}`))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"author_not_found"`) {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestListArticlesSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now().UTC().Truncate(time.Microsecond)
	store := &articleStoreStub{
		listReturn: []domain.Article{
			{ID: "3fa85f64-5717-4562-b3fc-2c963f66afa6", AuthorID: "550e8400-e29b-41d4-a716-446655440000", Title: "Title", Body: "Body", CreatedAt: now},
			{ID: "2fa85f64-5717-4562-b3fc-2c963f66afa6", AuthorID: "550e8400-e29b-41d4-a716-446655440000", Title: "Older Title", Body: "Older Body", CreatedAt: now},
		},
	}
	router := gin.New()
	ctrl := NewArticleController(usecase.NewArticleUsecase(store))
	router.GET("/articles", ctrl.List)

	req := httptest.NewRequest(http.MethodGet, "/articles?query=title&author=alice&limit=1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var resp ListArticlesResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Items) != 1 || resp.Meta.NextCursor == nil {
		t.Fatalf("unexpected resp: %#v", resp)
	}
}

func TestListArticlesInvalidLimitAndCursor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &articleStoreStub{}
	router := gin.New()
	ctrl := NewArticleController(usecase.NewArticleUsecase(store))
	router.GET("/articles", ctrl.List)

	req := httptest.NewRequest(http.MethodGet, "/articles?limit=invalid", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), `"invalid_limit"`) {
		t.Fatalf("invalid limit resp = %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/articles?cursor=bad", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), `"invalid_cursor"`) {
		t.Fatalf("invalid cursor resp = %d %s", rec.Code, rec.Body.String())
	}
}
