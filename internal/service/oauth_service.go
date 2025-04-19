package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/whitewalker-sa/ehass/internal/model"
)

// oauthService implements the OAuthService interface
type oauthService struct {
	githubClientID     string
	githubClientSecret string
	googleClientID     string
	googleClientSecret string
	httpClient         *http.Client
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(
	githubClientID string,
	githubClientSecret string,
	googleClientID string,
	googleClientSecret string,
) OAuthService {
	return &oauthService{
		githubClientID:     githubClientID,
		githubClientSecret: githubClientSecret,
		googleClientID:     googleClientID,
		googleClientSecret: googleClientSecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetUserInfo gets user information from OAuth provider
func (s *oauthService) GetUserInfo(ctx context.Context, provider model.AuthProvider, token string) (*OAuthUserInfo, error) {
	switch provider {
	case model.AuthProviderGithub:
		return s.getGithubUserInfo(ctx, token)
	case model.AuthProviderGoogle:
		return s.getGoogleUserInfo(ctx, token)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// getGithubUserInfo retrieves user information from GitHub
func (s *oauthService) getGithubUserInfo(ctx context.Context, token string) (*OAuthUserInfo, error) {
	// Create request to GitHub API
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	// Set authorization header
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	// Make request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned non-200 status code: %d", resp.StatusCode)
	}

	// Parse response
	var githubUser struct {
		ID        int    `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return nil, err
	}

	// If email is not provided, fetch user emails
	if githubUser.Email == "" {
		email, err := s.getGithubUserEmail(ctx, token)
		if err != nil {
			return nil, err
		}
		githubUser.Email = email
	}

	// Use login as name if name is not provided
	name := githubUser.Name
	if name == "" {
		name = githubUser.Login
	}

	return &OAuthUserInfo{
		ID:     fmt.Sprintf("%d", githubUser.ID),
		Email:  githubUser.Email,
		Name:   name,
		Avatar: githubUser.AvatarURL,
	}, nil
}

// getGithubUserEmail retrieves primary email from GitHub
func (s *oauthService) getGithubUserEmail(ctx context.Context, token string) (string, error) {
	// Create request to GitHub API
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}

	// Set authorization header
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	// Make request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned non-200 status code: %d", resp.StatusCode)
	}

	// Parse response
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	// Find primary and verified email
	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	// If no primary and verified email, use the first verified email
	for _, email := range emails {
		if email.Verified {
			return email.Email, nil
		}
	}

	return "", fmt.Errorf("no verified email found")
}

// getGoogleUserInfo retrieves user information from Google
func (s *oauthService) getGoogleUserInfo(ctx context.Context, token string) (*OAuthUserInfo, error) {
	// Create request to Google API
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, err
	}

	// Set authorization header
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	// Make request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google API returned non-200 status code: %d", resp.StatusCode)
	}

	// Parse response
	var googleUser struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, err
	}

	// Ensure we have an email
	if googleUser.Email == "" {
		return nil, fmt.Errorf("no email provided by Google")
	}

	// Use given name as name if name is not provided
	name := googleUser.Name
	if name == "" {
		name = googleUser.GivenName
		if googleUser.FamilyName != "" {
			name += " " + googleUser.FamilyName
		}
	}

	return &OAuthUserInfo{
		ID:     googleUser.Sub,
		Email:  googleUser.Email,
		Name:   name,
		Avatar: googleUser.Picture,
	}, nil
}
