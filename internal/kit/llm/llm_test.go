package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"charm.land/fantasy"
)

func clear(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv(EnvModel, "")
	t.Setenv(EnvBaseURL, "")
	for _, k := range keyed {
		t.Setenv(k.env, "")
	}
}

func TestAnEndpointOfOnesOwnIsAsked(t *testing.T) {
	clear(t)
	var asked struct {
		path, key, model string
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct{ Model string }
		json.NewDecoder(r.Body).Decode(&body)
		asked.path, asked.key, asked.model = r.URL.Path, r.Header.Get("Authorization"), body.Model
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"1","object":"chat.completion","created":1,"model":"local","choices":[{"index":0,"finish_reason":"stop","message":{"role":"assistant","content":"hello back"}}],"usage":{"prompt_tokens":1,"completion_tokens":2,"total_tokens":3}}`))
	}))
	defer srv.Close()
	t.Setenv(EnvBaseURL, srv.URL+"/v1")
	t.Setenv("OPENAI_API_KEY", "secret")
	t.Setenv(EnvModel, "local")
	// A key for another provider does not win over the endpoint.
	t.Setenv("ANTHROPIC_API_KEY", "other")

	model, err := FromEnv(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	res, err := fantasy.NewAgent(model).Generate(context.Background(), fantasy.AgentCall{Prompt: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if got := res.Response.Content.Text(); got != "hello back" {
		t.Errorf("answered %q", got)
	}
	if asked.path != "/v1/chat/completions" || asked.key != "Bearer secret" || asked.model != "local" {
		t.Errorf("asked as %+v", asked)
	}
}

func TestTheFirstKeySetNamesTheProvider(t *testing.T) {
	clear(t)
	if _, err := FromEnv(context.Background()); err == nil || !strings.Contains(err.Error(), EnvModel) {
		t.Errorf("with no model named: %v", err)
	}
	t.Setenv(EnvModel, "some-model")
	if _, err := FromEnv(context.Background()); err == nil || !strings.Contains(err.Error(), "ANTHROPIC_API_KEY") {
		t.Errorf("with no provider: %v", err)
	}
	t.Setenv("OPENROUTER_API_KEY", "b")
	t.Setenv("OPENAI_API_KEY", "a")
	model, err := FromEnv(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if model.Provider() != "openai" || model.Model() != "some-model" {
		t.Errorf("got %s of %s", model.Model(), model.Provider())
	}
}
