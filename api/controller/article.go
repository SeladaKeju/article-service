package controller

import (
	"encoding/json"
	"errors"
	"io"
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

type createArticleRequest struct {
	AuthorID *string `json:"author_id"`
	Title    *string `json:"title"`
	Body     *string `json:"body"`
}

type articleResponse struct {
	ID        string `json:"id"`
	AuthorID  string `json:"author_id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

func (a ArticleController) Create(c *gin.Context) {
	request, err := decodeCreateArticle(c.Request)
	if err != nil || request.AuthorID == nil || request.Title == nil || request.Body == nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "request must contain author_id, title, and body")
		return
	}

	article, err := a.usecase.Create(c.Request.Context(), usecase.CreateArticleInput{
		AuthorID: *request.AuthorID,
		Title:    *request.Title,
		Body:     *request.Body,
	})
	if errors.Is(err, usecase.ErrInvalidRequest) {
		writeError(c, http.StatusBadRequest, "invalid_request", "author_id, title, and body must not be blank")
		return
	}
	if usecase.IsAuthorNotFound(err) {
		writeError(c, http.StatusBadRequest, "author_not_found", "author was not found")
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "internal_error", "An internal error occurred")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": presentArticle(article)})
}

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

	data := make([]articleResponse, 0, len(result.Articles))
	for _, item := range result.Articles {
		data = append(data, presentArticle(item))
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        data,
		"next_cursor": result.NextCursor,
	})
}

func decodeCreateArticle(request *http.Request) (createArticleRequest, error) {
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	var value createArticleRequest
	if err := decoder.Decode(&value); err != nil {
		return createArticleRequest{}, err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return createArticleRequest{}, errors.New("request must contain one JSON object")
	}
	return value, nil
}

func presentArticle(article domain.Article) articleResponse {
	return articleResponse{
		ID:        article.ID,
		AuthorID:  article.AuthorID,
		Title:     article.Title,
		Body:      article.Body,
		CreatedAt: article.CreatedAt.UTC().Format("2006-01-02T15:04:05.000000Z"),
	}
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
