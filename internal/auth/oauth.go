package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// ─── Anthropic OAuth constants ────────────────────────────────────────────────
// These are the exact values extracted from the Claude Code CLI binary (v2.1.76).
// Verified by inspecting the minified JS bundle in the Bun-compiled binary.

const (
	// Authorization endpoint on claude.ai (CLAUDE_AI_AUTHORIZE_URL)
	anthropicAuthURL = "https://claude.ai/oauth/authorize"

	// Token endpoint on platform.claude.com (TOKEN_URL)
	anthropicTokenURL = "https://platform.claude.com/v1/oauth/token"

	// Claude Code's public client ID — extracted from the production config
	// block in the CLI binary. Public clients use PKCE instead of a secret.
	// Can be overridden at runtime via CLAUDE_CODE_OAUTH_CLIENT_ID env var.
	claudeCodeClientID = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"

	// CLAUDE_AI_OAUTH_SCOPES — exact scopes used by the Claude Code CLI for
	// claude.ai login. No OIDC scopes (openid/profile/email) — Anthropic uses
	// custom user: and org: scope prefixes.
	claudeCodeScope = "user:profile user:inference user:sessions:claude_code user:mcp_servers user:file_upload"

	// Required beta header for OAuth endpoints (OAUTH_BETA_HEADER)
	oauthBetaHeader = "oauth-2025-04-20"

	// Local callback server — registered as allowed redirect URI for public clients.
	callbackPort = 38787
	callbackPath = "/callback"
)

// OAuthResult holds the tokens returned after a successful OAuth flow.
type OAuthResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	TokenType    string
}

// OAuthFlow manages a single OAuth authorization code + PKCE flow.
type OAuthFlow struct {
	State        string
	Verifier     string
	Challenge    string
	AuthURL      string
	CallbackURL  string
	resultCh     chan OAuthResult
	errCh        chan error
	server       *http.Server
}

// NewOAuthFlow initializes a new PKCE OAuth flow for Claude Code.
func NewOAuthFlow() (*OAuthFlow, error) {
	state, err := randomHex(32)
	if err != nil {
		return nil, fmt.Errorf("generating state: %w", err)
	}

	verifier, challenge, err := pkce()
	if err != nil {
		return nil, fmt.Errorf("generating PKCE: %w", err)
	}

	callbackURL := fmt.Sprintf("http://127.0.0.1:%d%s", callbackPort, callbackPath)

	params := url.Values{
		"client_id":             {claudeCodeClientID},
		"redirect_uri":          {callbackURL},
		"response_type":         {"code"},
		"scope":                 {claudeCodeScope},
		"state":                 {state},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}

	authURL := anthropicAuthURL + "?" + params.Encode()

	return &OAuthFlow{
		State:       state,
		Verifier:    verifier,
		Challenge:   challenge,
		AuthURL:     authURL,
		CallbackURL: callbackURL,
		resultCh:    make(chan OAuthResult, 1),
		errCh:       make(chan error, 1),
	}, nil
}

// Start opens the browser and starts the local callback HTTP server.
// Returns immediately — use Wait() to block until the flow completes.
func (f *OAuthFlow) Start() error {
	// Start local callback server first
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", callbackPort))
	if err != nil {
		return fmt.Errorf("could not start callback server on port %d: %w\n"+
			"Try closing any other application using that port.", callbackPort, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, f.handleCallback)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, callbackPath, http.StatusFound)
	})

	f.server = &http.Server{
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		_ = f.server.Serve(listener)
	}()

	// Open browser
	return openURL(f.AuthURL)
}

// Wait blocks until the OAuth flow completes (user authorized or timed out).
// Timeout: 5 minutes.
func (f *OAuthFlow) Wait(ctx context.Context) (OAuthResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	defer f.shutdown()

	select {
	case result := <-f.resultCh:
		return result, nil
	case err := <-f.errCh:
		return OAuthResult{}, err
	case <-ctx.Done():
		return OAuthResult{}, fmt.Errorf("authentication timed out after 5 minutes")
	}
}

// handleCallback handles the OAuth redirect from Claude.
func (f *OAuthFlow) handleCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// Validate state (CSRF protection)
	if q.Get("state") != f.State {
		f.errCh <- fmt.Errorf("state mismatch — possible CSRF attack; please try again")
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Handle error response from provider
	if errParam := q.Get("error"); errParam != "" {
		desc := q.Get("error_description")
		f.errCh <- fmt.Errorf("authorization denied: %s — %s", errParam, desc)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, successPageHTML("Authorization denied. You can close this window.", false))
		return
	}

	code := q.Get("code")
	if code == "" {
		f.errCh <- fmt.Errorf("no authorization code received")
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	result, err := f.exchangeCode(r.Context(), code)
	if err != nil {
		f.errCh <- fmt.Errorf("token exchange failed: %w", err)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, successPageHTML("Authentication failed. You can close this window.", false))
		return
	}

	// Success
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, successPageHTML("Connected! You can close this window and return to Research Loop.", true))

	f.resultCh <- result
}

// exchangeCode POSTs the authorization code + PKCE verifier to the token endpoint.
func (f *OAuthFlow) exchangeCode(ctx context.Context, code string) (OAuthResult, error) {
	body := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {f.CallbackURL},
		"client_id":     {claudeCodeClientID},
		"code_verifier": {f.Verifier},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", anthropicTokenURL,
		strings.NewReader(body.Encode()))
	if err != nil {
		return OAuthResult{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Anthropic-Beta", oauthBetaHeader)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return OAuthResult{}, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return OAuthResult{}, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(raw))
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return OAuthResult{}, fmt.Errorf("parsing token response: %w", err)
	}
	if result.Error != "" {
		return OAuthResult{}, fmt.Errorf("%s: %s", result.Error, result.ErrorDesc)
	}

	return OAuthResult{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    result.TokenType,
	}, nil
}

func (f *OAuthFlow) shutdown() {
	if f.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = f.server.Shutdown(ctx)
	}
}

// ─── Token refresh ────────────────────────────────────────────────────────────

// RefreshAccessToken exchanges a refresh token for a new access token.
func RefreshAccessToken(refreshToken string) (OAuthResult, error) {
	body := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {claudeCodeClientID},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", anthropicTokenURL,
		strings.NewReader(body.Encode()))
	if err != nil {
		return OAuthResult{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Anthropic-Beta", oauthBetaHeader)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return OAuthResult{}, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return OAuthResult{}, fmt.Errorf("refresh failed (%d): %s", resp.StatusCode, string(raw))
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Error        string `json:"error"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return OAuthResult{}, err
	}
	if result.Error != "" {
		return OAuthResult{}, fmt.Errorf("refresh error: %s", result.Error)
	}

	return OAuthResult{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    result.TokenType,
	}, nil
}

// ─── Extended credential storage for OAuth tokens ────────────────────────────

// OAuthCredential extends Credential with OAuth-specific fields.
type OAuthCredential struct {
	Credential
	RefreshToken string
	ExpiresAt    time.Time
}

// SaveOAuth stores an OAuth result as a credential.
func SaveOAuth(workspaceRoot, providerID string, result OAuthResult) error {
	cred := Credential{
		ProviderID: providerID,
		Value:      result.AccessToken,
	}
	if err := Save(workspaceRoot, cred); err != nil {
		return err
	}
	// Also save refresh token separately
	if result.RefreshToken != "" {
		refreshCred := Credential{
			ProviderID: providerID + ":refresh",
			Value:      result.RefreshToken,
		}
		return Save(workspaceRoot, refreshCred)
	}
	return nil
}

// LoadOAuth returns the stored OAuth credential, refreshing if needed.
func LoadOAuth(workspaceRoot, providerID string) (string, error) {
	cred, ok := Load(workspaceRoot, providerID)
	if !ok {
		return "", fmt.Errorf("not authenticated with %s — run 'research-loop tui' and choose Setup Provider", providerID)
	}

	// Try refresh if we have a refresh token
	refreshCred, hasRefresh := Load(workspaceRoot, providerID+":refresh")
	if hasRefresh && refreshCred.Value != "" {
		newResult, err := RefreshAccessToken(refreshCred.Value)
		if err == nil {
			_ = SaveOAuth(workspaceRoot, providerID, newResult)
			return newResult.AccessToken, nil
		}
		// Refresh failed — use stored access token and hope it's still valid
	}

	return cred.Value, nil
}

// ─── PKCE helpers ─────────────────────────────────────────────────────────────

func pkce() (verifier, challenge string, err error) {
	b := make([]byte, 64)
	if _, err = rand.Read(b); err != nil {
		return
	}
	verifier = base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])
	return
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}

// ─── Browser opener ───────────────────────────────────────────────────────────

func openURL(rawURL string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", rawURL)
	case "linux":
		cmd = exec.Command("xdg-open", rawURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL)
	default:
		// Can't open browser — caller should display the URL
		fmt.Fprintf(os.Stderr, "\nOpen this URL in your browser:\n%s\n\n", rawURL)
		return nil
	}
	return cmd.Start()
}

// ─── Callback page HTML ───────────────────────────────────────────────────────

func successPageHTML(message string, success bool) string {
	icon := "✓"
	color := "#22c55e"
	if !success {
		icon = "✗"
		color = "#ef4444"
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Research Loop</title>
  <style>
    body { font-family: -apple-system, sans-serif; background: #0f0f0f; color: #e0e0e0;
           display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }
    .card { background: #1a1a1a; border: 1px solid #2a2a2a; border-radius: 12px;
            padding: 48px; text-align: center; max-width: 400px; }
    .icon { font-size: 48px; color: %s; margin-bottom: 16px; }
    h1 { font-size: 20px; font-weight: 600; margin: 0 0 8px; }
    p { color: #888; font-size: 14px; margin: 0; }
    .logo { font-size: 12px; color: #444; margin-top: 24px; letter-spacing: 0.1em; }
  </style>
</head>
<body>
  <div class="card">
    <div class="icon">%s</div>
    <h1>%s</h1>
    <p>You can close this tab and return to your terminal.</p>
    <div class="logo">🔬 RESEARCH LOOP</div>
  </div>
  <script>setTimeout(() => window.close(), 3000)</script>
</body>
</html>`, color, icon, message)
}
