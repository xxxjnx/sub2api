package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type usageDetailRepoStub struct {
	service.UsageLogRepository
	record          *service.UsageLog
	requestedID     int64
	requestedUserID int64
}

func (s *usageDetailRepoStub) GetByID(_ context.Context, id int64) (*service.UsageLog, error) {
	s.requestedID = id
	return s.record, nil
}

func (s *usageDetailRepoStub) GetByIDForUser(_ context.Context, id, userID int64) (*service.UsageLog, error) {
	s.requestedID = id
	s.requestedUserID = userID
	if s.record == nil || s.record.UserID != userID {
		return nil, service.ErrUsageLogNotFound
	}
	return s.record, nil
}

func newUsageDetailTestRouter(repo *usageDetailRepoStub, userID int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	usageService := service.NewUsageService(repo, nil, nil, nil, nil)
	usageHandler := NewUsageHandler(usageService, nil, nil, nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: userID})
		c.Next()
	})
	router.GET("/usage/:id", usageHandler.GetByID)
	return router
}

func TestUsageGetByIDNeverReturnsOriginalRequestDataToOwner(t *testing.T) {
	contentType := "application/json"
	repo := &usageDetailRepoStub{record: &service.UsageLog{
		ID:                 42,
		UserID:             7,
		APIKeyID:           9,
		AccountID:          11,
		RequestID:          "req-user-detail-42",
		Model:              "gpt-5",
		RequestData:        []byte(`{"authorization":"Bearer raw-owner-secret","api_key":"sk-visible"}`),
		RequestContentType: &contentType,
	}}
	router := newUsageDetailTestRouter(repo, 7)

	req := httptest.NewRequest(http.MethodGet, "/usage/42", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
	require.Equal(t, int64(42), repo.requestedID)
	require.Equal(t, int64(7), repo.requestedUserID)
	require.NotContains(t, rec.Body.String(), "raw-owner-secret")
	require.NotContains(t, rec.Body.String(), "sk-visible")
	require.NotContains(t, rec.Body.String(), `"request_data"`)
	require.NotContains(t, rec.Body.String(), `"request_content_type"`)
}

func TestUsageGetByIDDoesNotExposeRequestDataAcrossUsers(t *testing.T) {
	repo := &usageDetailRepoStub{record: &service.UsageLog{
		ID:          42,
		UserID:      99,
		RequestData: []byte(`{"authorization":"Bearer other-user-secret"}`),
	}}
	router := newUsageDetailTestRouter(repo, 7)

	req := httptest.NewRequest(http.MethodGet, "/usage/42", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Equal(t, int64(7), repo.requestedUserID)
	require.NotContains(t, rec.Body.String(), "other-user-secret")
}
