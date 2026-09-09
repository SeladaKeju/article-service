package controller

import (
	"errors"
	"net/http"

	"github.com/SeladaKeju/article-service.git/domain"
	"github.com/SeladaKeju/article-service.git/usecase"
	"github.com/gin-gonic/gin"
)

type ArticleController struct {
	usecase *usecase.ArticleUsecase
}

func NewArticleController(usecase *usecase.ArticleUsecase) ArticleController {
	return ArticleController{usecase: usecase}
}

type CreateArticleRequest struct {
	AuthorID *string `json:"author_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title    *string `json:"title" binding:"required" example:"Go Concurrency"`
	Body     *string `json:"body" binding:"required" example:"Concurrent requests in Go."`
}

type ArticleResponse struct {
	ID        string `json:"id"`
	AuthorID  string `json:"author_id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

type ArticleEnvelope struct {
	Data ArticleResponse `json:"data"`
}

type ArticlesEnvelope struct {
	Data       []ArticleResponse `json:"data"`
	NextCursor *string           `json:"next_cursor" example:"eyJjcmVhdGVkX2F0IjoiMjAyNi0wOS0wOFQxMjowMDowMC4xMjM0NTZaIiwiaWQiOiIzZmE4NWY2NC01NzE3LTQ1NjItYjNmYy0yYzk2M2Y2NmFmYTYifQ"`
}

type ErrorDetail struct {
	Code    string `json:"code" example:"invalid_request"`
	Message string `json:"message" example:"request must contain author_id, title, and body"`
}

type ErrorEnvelope struct {
	Error ErrorDetail `json:"error"`
}

// Create creates an article for an existing author.
// @Summary Create an article
// @Tags articles
// @Accept json
// @Produce json
// @Param article body CreateArticleRequest true "Article payload"
// @Success 201 {object} ArticleEnvelope
// @Failure 400 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /articles [post]
func (a ArticleController) Create(c *gin.Context) {
	var request CreateArticleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "request must contain author_id, title, and body")
		return
	}

	article, err := a.usecase.Create(c.Request.Context(), usecase.CreateArticleInput{
		AuthorID: *request.AuthorID,
		Title:    *request.Title,
		Body:     *request.Body,
	})
	if errors.Is(err, domain.ErrInvalidArticle) {
		writeError(c, http.StatusBadRequest, "invalid_request", "author_id, title, and body must not be blank")
		return
	}
	if errors.Is(err, domain.ErrAuthorNotFound) {
		writeError(c, http.StatusBadRequest, "author_not_found", "author was not found")
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "internal_error", "An internal error occurred")
		return
	}

	c.JSON(http.StatusCreated, ArticleEnvelope{Data: presentArticle(article)})
}

// List returns articles newest first, optionally filtered by keyword and author.
// @Summary List articles
// @Tags articles
// @Produce json
// @Param query query string false "Whole-word search across title and body"
// @Param author query string false "Full author name, case-insensitive"
// @Param limit query int false "Results per page (1-100, default 20)"
// @Param cursor query string false "Cursor returned by the previous page"
// @Success 200 {object} ArticlesEnvelope
// @Failure 400 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /articles [get]
func (a ArticleController) List(c *gin.Context) {
	result, err := a.usecase.List(c.Request.Context(), usecase.ListArticlesInput{
		Query:  c.Query("query"),
		Author: c.Query("author"),
		Limit:  c.Query("limit"),
		Cursor: c.Query("cursor"),
	})
	if errors.Is(err, usecase.ErrInvalidLimit) {
		writeError(c, http.StatusBadRequest, "invalid_limit", "limit must be an integer between 1 and 100")
		return
	}
	if errors.Is(err, usecase.ErrInvalidCursor) {
		writeError(c, http.StatusBadRequest, "invalid_cursor", "cursor is invalid")
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "internal_error", "An internal error occurred")
		return
	}

	data := make([]ArticleResponse, 0, len(result.Articles))
	for _, item := range result.Articles {
		data = append(data, presentArticle(item))
	}

	c.JSON(http.StatusOK, ArticlesEnvelope{Data: data, NextCursor: result.NextCursor})
}

func presentArticle(article domain.Article) ArticleResponse {
	return ArticleResponse{
		ID:        article.ID,
		AuthorID:  article.AuthorID,
		Title:     article.Title,
		Body:      article.Body,
		CreatedAt: article.CreatedAt.UTC().Format("2006-01-02T15:04:05.000000Z"),
	}
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorEnvelope{Error: ErrorDetail{Code: code, Message: message}})
}
