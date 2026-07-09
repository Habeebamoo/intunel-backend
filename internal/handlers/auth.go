package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Habeebamoo/intunel-backend/internal/configs"
	"github.com/Habeebamoo/intunel-backend/internal/models"
	"github.com/Habeebamoo/intunel-backend/internal/services"
	"github.com/Habeebamoo/intunel-backend/internal/store"
	"github.com/Habeebamoo/intunel-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type AuthHandler struct {
	service      services.AuthService
	googleConfig *oauth2.Config
	githubConfig *oauth2.Config
	stateStore   *store.OAuthStateStore
	frontendURL  string
	cfg          *configs.Config
}

func NewAuthHandler(s services.AuthService, stateStore *store.OAuthStateStore, cfg *configs.Config) *AuthHandler {
	googleConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  "http://localhost:8080/api/v1/auth/google/callback",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	githubConfig := &oauth2.Config{
		ClientID:     cfg.GitHubClientID,
		ClientSecret: cfg.GitHubClientSecret,
		RedirectURL:  "http://localhost:8080/api/v1/auth/github/callback",
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	return &AuthHandler{
		service:      s,
		googleConfig: googleConfig,
		githubConfig: githubConfig,
		stateStore:   stateStore,
		frontendURL:  cfg.FrontendUrl,
		cfg:          cfg,
	}
}

// Google
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	state, err := h.stateStore.Generate(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to generate state")
		return
	}
	url := h.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	state := c.Query("state")
	if err := h.stateStore.Verify(c.Request.Context(), state); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid state parameter")
		return
	}

	code := c.Query("code")
	token, err := h.googleConfig.Exchange(context.Background(), code)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "failed to exchange token")
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to fetch user info")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var googleUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	json.Unmarshal(body, &googleUser)

	oauthUser := &models.OAuthUser{
		Name:   googleUser.Name,
		Email:  googleUser.Email,
		Avatar: googleUser.Picture,
	}

	authResp, err := h.service.HandleOAuth(c.Request.Context(), oauthUser)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.setAuthCookie(c, authResp.Token)
	c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/dashboard")
}

// GitHub
func (h *AuthHandler) GitHubLogin(c *gin.Context) {
	state, err := h.stateStore.Generate(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to generate state")
		return
	}
	url := h.githubConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GitHubCallback(c *gin.Context) {
	state := c.Query("state")
	if err := h.stateStore.Verify(c.Request.Context(), state); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid state parameter")
		return
	}

	code := c.Query("code")
	token, err := h.githubConfig.Exchange(context.Background(), code)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "failed to exchange token")
		return
	}

	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to fetch user info")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var githubUser struct {
		ID        int    `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	json.Unmarshal(body, &githubUser)

	email := githubUser.Email
	if email == "" {
		email = h.fetchGitHubEmail(token.AccessToken)
	}

	oauthUser := &models.OAuthUser{
		Name:   githubUser.Name,
		Email:  email,
		Avatar: githubUser.AvatarURL,
	}

	authResp, err := h.service.HandleOAuth(c.Request.Context(), oauthUser)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.setAuthCookie(c, authResp.Token)
	c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/dashboard")
}

func (h *AuthHandler) fetchGitHubEmail(accessToken string) string {
	req, _ := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}
	json.Unmarshal(body, &emails)

	for _, e := range emails {
		if e.Primary {
			return e.Email
		}
	}
	return ""
}

func (h *AuthHandler) setAuthCookie(c *gin.Context, token string) {
	c.SetCookie(
		"intunel_token",
		token,
		60*60*72,
		"/",
		"",
		h.cfg.Env == "production",
		true,
	)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("intunel_token", "", -1, "/", "", false, true)
	utils.SuccessResponse(c, http.StatusOK, "logged out", nil)
}