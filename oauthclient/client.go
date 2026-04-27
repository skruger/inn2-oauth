package oauthclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

type OauthClient struct {
	ClientID           string
	ClientSecret       string
	TokenURL           string
	IdentityURL        string
	UsernameFields     []string
	OauthTokenUsername string
	QuickAuth          bool
	QuickDomain        string
}

func NewOauthClient(cfg ClientConfig) *OauthClient {
	return &OauthClient{
		ClientID:           cfg.ClientId,
		ClientSecret:       cfg.ClientSecret,
		TokenURL:           cfg.TokenURL,
		IdentityURL:        cfg.IdentityURL,
		UsernameFields:     cfg.UsernameFields,
		OauthTokenUsername: cfg.OauthTokenUsername,
		QuickAuth:          cfg.QuickAuth,
		QuickDomain:        cfg.QuickDomain,
	}
}

// ObtainAccessToken sends an HTTP POST to the provided URL with the
// same form parameters as the curl command and returns the response body.
// urlStr should be the full URL, e.g. "http://localhost:8000/oauth2/access_token".
func (oc *OauthClient) ObtainAccessToken(username, password string) (*TokenResponse, error) {
	form := url.Values{}
	form.Set("client_id", oc.ClientID)
	form.Set("client_secret", oc.ClientSecret)
	form.Set("username", username)
	form.Set("password", password)
	form.Set("grant_type", "password")

	resp, err := http.PostForm(oc.TokenURL, form)
	if err != nil {
		return nil, fmt.Errorf("post form to %s with username %s: %w", oc.TokenURL, username, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// return body for debugging along with an error
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, body)
	}
	tokenResponse := TokenResponse{}
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return &tokenResponse, nil
}

func (oc *OauthClient) CheckIdentity(token string) (map[string]string, error) {
	var identityResponse map[string]string
	if token == "" {
		return identityResponse, fmt.Errorf("empty token")
	}
	req, err := http.NewRequest(http.MethodGet, oc.IdentityURL, nil)
	if err != nil {
		return identityResponse, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return identityResponse, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return identityResponse, fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return identityResponse, fmt.Errorf("identity check failed: status %d: %s", resp.StatusCode, string(body))
	}
	// Success
	err = json.Unmarshal(body, &identityResponse)

	return identityResponse, err
}
