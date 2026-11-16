package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	F "github.com/IBM/fp-go/v2/function"
	"github.com/dictybase/dcr-mcp/internal/server"
	"github.com/stretchr/testify/require"
)

const (
	challengePath  = "/.well-known/openai-apps-challenge"
	challengeToken = "test-domain-verification-value"
)

func challengeResponse(t *testing.T, token string) (int, string, string) {
	t.Helper()
	httpSrv := httptest.NewServer(server.NewHTTPHandler(F.Pipe1(
		server.Config{},
		server.ConfigOpenAIChallengeLens.Set(token),
	)))
	t.Cleanup(httpSrv.Close)
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, httpSrv.URL+challengePath, nil)
	require.NoError(t, err)
	res, err := httpSrv.Client().Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return res.StatusCode, res.Header.Get("Content-Type"), string(body)
}

func TestOpenAIChallengeServesToken(t *testing.T) {
	t.Parallel()
	status, contentType, body := challengeResponse(t, challengeToken)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, "text/plain; charset=utf-8", contentType)
	require.Equal(t, challengeToken, body)
}

func TestOpenAIChallengeDisabledWithoutToken(t *testing.T) {
	t.Parallel()
	status, _, _ := challengeResponse(t, "")
	require.Equal(t, http.StatusNotFound, status)
}
