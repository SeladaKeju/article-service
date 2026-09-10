package controller

import (
	"errors"
	"net/http"

	"github.com/SeladaKeju/article-service.git/domain"
	"github.com/gin-gonic/gin"
)

// ArticleController translates HTTP requests to article use-case calls.
type ArticleController struct {
	usecase domain.ArticleUsecase
}

// NewArticleController constructs an article HTTP controller.
func NewArticleController(usecase domain.ArticleUsecase) ArticleController {
	return ArticleController{usecase: usecase}
}

// ListArticlesMeta contains pagination metadata for an article list.
type ListArticlesMeta struct {
	NextCursor *string `json:"next_cursor"`
	
}

// ListArticlesResponse contains an article page and its pagination metadata.
type ListArticlesResponse struct {
	Items []domain.Article `json:"items"`
	Meta  ListArticlesMeta `json:"meta"`
}

// ErrorDetail describes an API error.
type ErrorDetail struct {
	Code    string `json:"code" example:"invalid_request"`
	Message string `json:"message" example:"request must contain author_id, title, and body"`
}

// Create creates an article for an existing author.
// @Summary Create an article
// @Tags articles
// @Accept json
// @Produce json
// @Param article body domain.Article true "Article payload"
// @Success 201 {object} domain.Article
// @Failure 400 {object} ErrorDetail
// @Failure 500 {object} ErrorDetail
// @Router /articles [post]
func (a ArticleController) Create(c *gin.Context) {
	var article domain.Article
	if err := c.ShouldBindJSON(&article); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "request must contain author_id, title, and body")
		return
	}

	article, err := a.usecase.Create(c.Request.Context(), article)
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

	c.JSON(http.StatusCreated, article)
}

// List returns articles newest first, optionally filtered by keyword and author.
// @Summary List articles
// @Tags articles
// @Produce json
// @Param query query string false "Whole-word search across title and body"
// @Param author query string false "Full author name, case-insensitive"
// @Param limit query int false "Results per page (1-100, default 20)"
// @Param cursor query string false "Cursor returned by the previous page"
// @Success 200 {object} ListArticlesResponse
// @Failure 400 {object} ErrorDetail
// @Failure 500 {object} ErrorDetail
// @Router /articles [get]
func (a ArticleController) List(c *gin.Context) {
	result, err := a.usecase.List(c.Request.Context(), domain.ListArticlesInput{
		Query:  c.Query("query"),
		Author: c.Query("author"),
		Limit:  c.Query("limit"),
		Cursor: c.Query("cursor"),
	})
	if errors.Is(err, domain.ErrInvalidLimit) {
		writeError(c, http.StatusBadRequest, "invalid_limit", "limit must be an integer between 1 and 100")
		return
	}
	if errors.Is(err, domain.ErrInvalidCursor) {
		writeError(c, http.StatusBadRequest, "invalid_cursor", "cursor is invalid")
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "internal_error", "An internal error occurred")
		return
	}

	c.JSON(http.StatusOK, ListArticlesResponse{
		Items: result.Articles,
		Meta:  ListArticlesMeta{NextCursor: result.NextCursor},
	})
}

// writeError writes the service's standard error response.
func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorDetail{Code: code, Message: message})
}
