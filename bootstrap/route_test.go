package bootstrap

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"seanmcapp/repository"
	"seanmcapp/service"
	"seanmcapp/util"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		{"success", nil, http.StatusOK},
		{"validation", service.ValidationError{Message: "bad"}, http.StatusBadRequest},
		{"not found", repository.ErrNotFound, http.StatusNotFound},
		{"internal", errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			resolve(c, "payload", tc.err)
			assert.Equal(t, tc.wantCode, w.Code)
			if tc.err == nil {
				assert.Contains(t, w.Body.String(), "payload")
			}
		})
	}
}

type sampleReq struct {
	Value int `json:"value"`
}

func TestHandleJSON(t *testing.T) {
	r := gin.New()
	r.POST("/t", handleJSON(func(req sampleReq) (int, error) {
		return req.Value * 2, nil
	}))

	t.Run("valid body", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/t", strings.NewReader(`{"value":21}`))
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "42")
	})

	t.Run("invalid body", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/t", strings.NewReader(`not-json`))
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthMiddleware(t *testing.T) {
	settings := util.WalletSettings{SecretKey: "secret", Password: "pw"}
	token := util.JwtCreateToken(settings, "pw")
	require.NotEmpty(t, token)

	r := gin.New()
	r.GET("/protected", authMiddleware(settings), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	t.Run("valid token passes", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", token)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid token rejected", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer bad")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("preflight OPTIONS passes through", func(t *testing.T) {
		// Invoke the middleware directly: gin routing wouldn't match OPTIONS to a GET route.
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodOptions, "/protected", nil)
		authMiddleware(settings)(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestServeIndexMissingFile(t *testing.T) {
	// The frontend build isn't present in the test working directory.
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	serveIndex(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

