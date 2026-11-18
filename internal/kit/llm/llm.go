// Package llm gives the language model the environment names, or failing
// that the user's config file. It asks nothing of a model itself.
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"charm.land/fantasy"
	"charm.land/fantasy/providers/anthropic"
	"charm.land/fantasy/providers/google"
	"charm.land/fantasy/providers/openai"
	"charm.land/fantasy/providers/openaicompat"
	"charm.land/fantasy/providers/openrouter"
)

// The variables read. Fantasy reads none itself.
const (
	EnvModel   = "LGTM_MODEL"      // the model's id, as its provider writes it
	EnvBaseURL = "OPENAI_BASE_URL" // an endpoint that speaks OpenAI's API
	EnvEffort  = "LGTM_EFFORT"     // how hard the model thinks before it answers: none, minimal, low, medium or high; low when unset
)

// keyed is the providers that need only a key, in the order they are tried.
var keyed = []struct {
	env string
	new func(key string) (fantasy.Provider, error)
}{
	{"ANTHROPIC_API_KEY", func(k string) (fantasy.Provider, error) { return anthropic.New(anthropic.WithAPIKey(k)) }},
	{"OPENAI_API_KEY", func(k string) (fantasy.Provider, error) { return openai.New(openai.WithAPIKey(k)) }},
	{"OPENROUTER_API_KEY", func(k string) (fantasy.Provider, error) { return openrouter.New(openrouter.WithAPIKey(k)) }},
	{"GEMINI_API_KEY", func(k string) (fantasy.Provider, error) { return google.New(google.WithGeminiAPIKey(k)) }},
}

// Config is the file, under the home directory, that holds what the
// environment does not: one JSON object, its keys the names of the
// variables.
//
//	{"LGTM_MODEL": "…", "OPENAI_BASE_URL": "…", "OPENAI_API_KEY": "…"}
const Config = ".lgtm/config.json"

// named reads a variable: from the environment, or else from the config.
var named = func() func(string) string {
	var file map[string]string
	read := false
	return func(name string) string {
		if v := os.Getenv(name); v != "" {
			return v
		}
		if !read {
			read = true
			if home, err := os.UserHomeDir(); err == nil {
				if src, err := os.ReadFile(filepath.Join(home, Config)); err == nil {
					json.Unmarshal(src, &file)
				}
			}
		}
		return file[name]
	}
}()

// FromEnv gives the model the environment names. With OPENAI_BASE_URL set
// it is whatever answers there, with OPENAI_API_KEY as its key if it wants
// one. Otherwise it is the first provider whose key is set. What the
// environment does not set is read from the config file.
func FromEnv(ctx context.Context) (fantasy.LanguageModel, error) {
	id := named(EnvModel)
	if id == "" {
		return nil, fmt.Errorf("%s is not set, here or in ~/%s: it names the model to ask", EnvModel, Config)
	}
	p, err := provider()
	if err != nil {
		return nil, err
	}
	return p.LanguageModel(ctx, id)
}

// Options gives what the environment asks of every call to the model. So
// far that is how hard it thinks, and only an endpoint of one's own is
// told: the others keep their own default.
func Options() fantasy.ProviderOptions {
	if named(EnvBaseURL) == "" {
		return nil
	}
	// Low is where the reader waits a quarter less and the document reads
	// the same; none saves little more.
	effort := named(EnvEffort)
	if effort == "" {
		effort = "low"
	}
	return openaicompat.NewProviderOptions(&openaicompat.ProviderOptions{
		ReasoningEffort: openai.ReasoningEffortOption(openai.ReasoningEffort(effort)),
	})
}

func provider() (fantasy.Provider, error) {
	if url := named(EnvBaseURL); url != "" {
		return openaicompat.New(openaicompat.WithBaseURL(url), openaicompat.WithAPIKey(named("OPENAI_API_KEY")))
	}
	names := EnvBaseURL
	for _, k := range keyed {
		if key := named(k.env); key != "" {
			return k.new(key)
		}
		names += ", " + k.env
	}
	return nil, fmt.Errorf("no provider: set one of %s, here or in ~/%s", names, Config)
}
