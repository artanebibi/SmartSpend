package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var (
	router = gin.Default()
)

func TestHealthCheck(t *testing.T) {
	// Assume you have a route like this in main.go
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
}

func TestSignInWithGoogle(t *testing.T) {
	router.POST("/auth/google/signin", func(c *gin.Context) {

	})
}

func TestGoogleSignIn_Live(t *testing.T) {
	validGoogleIDToken := os.Getenv("GOOGLEID_TOKEN")

	payload := map[string]string{
		"id_token": validGoogleIDToken,
	}
	body, _ := json.Marshal(payload)

	URL := fmt.Sprintf("%s/api/auth/signin/google", os.Getenv("BASE_URL"))

	resp, err := http.Post(URL, "application/json", bytes.NewBuffer(body))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
