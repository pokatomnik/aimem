package main

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/caarlos0/env/v11"
)

// Config is parsed once, by caarlos0/env, from the process environment, and a
// tag option it does not understand is not a harmless typo: `get` rejects the
// option before it looks at a single variable, so Parse fails for every
// deployment and `scaffold` turns that into log.Fatalf. Nothing caught that
// until this test, and the failure it guards against was real — `notEmpty` was
// rejected by the version pinned at the time.
//
// Only the "option not supported" error is asserted on. A field marked
// `required` also fails the probe, because the probe reads the process
// environment; that outcome is expected here and says nothing about the tag.
func TestEveryConfigEnvTagIsAcceptedByTheParser(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(Config{})
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		tag, ok := field.Tag.Lookup("env")
		if !ok {
			t.Fatalf("Config.%s has no env tag", field.Name)
		}

		// A one-field struct carrying the same tag, so the parser sees the tag
		// without the rest of Config and without a value to hydrate.
		probe := reflect.New(reflect.StructOf([]reflect.StructField{{
			Name: "Probe",
			Type: reflect.TypeOf(""),
			Tag:  reflect.StructTag(fmt.Sprintf(`env:%q`, tag)),
		}}))

		err := env.Parse(probe.Interface())
		if err != nil && strings.Contains(err.Error(), "not supported") {
			t.Errorf("Config.%s tag %q: %v", field.Name, tag, err)
		}
	}
}

// Every runtime setting is reached through the AIMEM_ prefix; the unprefixed
// LLM_HOST of an earlier revision was read from nowhere and documented
// everywhere, which is the mistake this guards against.
func TestEveryConfigEnvNameUsesTheAIMEMPrefix(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(Config{})
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		name, _, _ := strings.Cut(field.Tag.Get("env"), ",")

		if !strings.HasPrefix(name, "AIMEM_") {
			t.Errorf("Config.%s reads %q, want an AIMEM_ name", field.Name, name)
		}
	}
}

// The tags are the contract that README.md documents: each name, whether a
// value is mandatory, and its default. Pinning the whole struct here means a
// renamed or re-defaulted setting cannot ship without the docs being revisited.
func TestConfigEnvTags(t *testing.T) {
	t.Parallel()

	want := map[string]string{
		"LLMHost":             "AIMEM_LLM_HOST,notEmpty",
		"LLMAPIKey":           "AIMEM_LLM_API_KEY",
		"SummarizationModel":  "AIMEM_SUMMARIZATION_MODEL,notEmpty",
		"TopK":                "AIMEM_TOP_K",
		"SimilarityThreshold": "AIMEM_SIMILARITY_THRESHOLD",
		"Dimensions":          "AIMEM_DIMENSIONS,notEmpty",
		"WebHTTPHost":         "AIMEM_WEB_HTTP_HOST",
		"QdrantHost":          "AIMEM_QDRANT_HOST",
		"QdrantAPIKey":        "AIMEM_QDRANT_API_KEY",
		"UsersJSON":           "AIMEM_USERS_JSON",
	}

	typ := reflect.TypeOf(Config{})
	if typ.NumField() != len(want) {
		t.Fatalf("Config has %d fields, the test knows %d: add the new one here", typ.NumField(), len(want))
	}

	for field, tag := range want {
		ref, ok := typ.FieldByName(field)
		if !ok {
			t.Fatalf("Config has no field %q", field)
		}

		if got := ref.Tag.Get("env"); got != tag {
			t.Errorf("Config.%s env tag = %q, want %q", field, got, tag)
		}
	}
}

// The defaults the docs promise, read straight from the tags.
func TestConfigEnvDefaults(t *testing.T) {
	t.Parallel()

	want := map[string]string{
		"TopK":                "5",
		"SimilarityThreshold": "0.7",
		"WebHTTPHost":         "127.0.0.1:8888",
	}

	typ := reflect.TypeOf(Config{})
	for field, wantDefault := range want {
		ref, ok := typ.FieldByName(field)
		if !ok {
			t.Fatalf("Config has no field %q", field)
		}

		if got := ref.Tag.Get("envDefault"); got != wantDefault {
			t.Errorf("Config.%s envDefault = %q, want %q", field, got, wantDefault)
		}
	}

	// Every other setting has no default: an absent Qdrant host is what selects
	// the embedded backend, and a default there would silently pick Qdrant.
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if _, ok := want[field.Name]; ok {
			continue
		}

		if _, ok := field.Tag.Lookup("envDefault"); ok {
			t.Errorf("Config.%s has an unexpected envDefault", field.Name)
		}
	}
}

func TestDimensionsUsesATypeSupportedByEnvParser(t *testing.T) {
	t.Parallel()

	field, ok := reflect.TypeOf(Config{}).FieldByName("Dimensions")
	if !ok {
		t.Fatal("Config has no Dimensions field")
	}
	if got := field.Type.Kind(); got != reflect.Uint {
		t.Fatalf("Config.Dimensions kind = %s, want uint supported by the env parser", got)
	}
}

func TestGetQdrantHostReturnsConfiguredHost(t *testing.T) {
	t.Parallel()

	const host = "http://qdrant.example:6333"
	config := Config{QdrantHost: host}

	if got := config.GetQdrantHost(); got != host {
		t.Fatalf("GetQdrantHost() = %q, want %q", got, host)
	}
}

func TestRequireAuth(t *testing.T) {
	t.Parallel()

	emptyConfig := Config{}
	if emptyConfig.RequireAuth() {
		t.Fatal("RequireAuth() = true for an empty host")
	}

	configuredConfig := Config{QdrantHost: "http://qdrant.example:6333"}
	if !configuredConfig.RequireAuth() {
		t.Fatal("RequireAuth() = false for a configured host")
	}
}

// minimalEnvironment is the smallest set that parses successfully: the three
// settings the tags mark as mandatory. The map reaches the parser through
// getConfig, so nothing is read from the process environment.
func minimalEnvironment() map[string]string {
	return map[string]string{
		"AIMEM_LLM_HOST":            "http://127.0.0.1:11434/v1",
		"AIMEM_SUMMARIZATION_MODEL": "gemma4:e2b",
		"AIMEM_DIMENSIONS":          "768",
	}
}

// Every setting reaches its getter, including the ones that are easy to wire to
// the wrong field.
func TestGetConfigReadsEverySetting(t *testing.T) {
	t.Parallel()

	environ := minimalEnvironment()
	environ["AIMEM_LLM_API_KEY"] = "secret"
	environ["AIMEM_TOP_K"] = "7"
	environ["AIMEM_SIMILARITY_THRESHOLD"] = "0.4"
	environ["AIMEM_WEB_HTTP_HOST"] = "0.0.0.0:9000"
	environ["AIMEM_QDRANT_HOST"] = "http://qdrant.example:6333"
	environ["AIMEM_QDRANT_API_KEY"] = "qk"
	environ["AIMEM_USERS_JSON"] = `[{"id":"alice","tokens":["t"]}]`

	cfg, err := getConfig(environ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := map[string]struct{ got, want any }{
		"LLMHost":             {cfg.GetLLMHost(), "http://127.0.0.1:11434/v1"},
		"LLMAPIKey":           {cfg.GetLLMAPIKey(), "secret"},
		"SummarizationModel":  {cfg.GetSummarizationModel(), "gemma4:e2b"},
		"TopK":                {cfg.GetTopK(), 7},
		"SimilarityThreshold": {cfg.GetSimilarityThreshold(), float32(0.4)},
		"Dimensions":          {cfg.GetDimensions(), uint32(768)},
		"WebHTTPHost":         {cfg.GetWebHTTPHost(), "0.0.0.0:9000"},
		"QdrantHost":          {cfg.GetQdrantHost(), "http://qdrant.example:6333"},
		"QdrantAPIKey":        {cfg.GetQdrantAPIKey(), "qk"},
		"UsersJSON":           {cfg.GetUsersJSON(), `[{"id":"alice","tokens":["t"]}]`},
		"RequireAuth":         {cfg.RequireAuth(), true},
	}

	for name, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %v, want %v", name, check.got, check.want)
		}
	}
}

// The defaults are what README.md promises for an otherwise minimal setup.
func TestGetConfigDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := getConfig(minimalEnvironment())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	switch {
	case cfg.GetTopK() != 5:
		t.Errorf("GetTopK() = %d, want 5", cfg.GetTopK())
	case cfg.GetSimilarityThreshold() != 0.7:
		t.Errorf("GetSimilarityThreshold() = %v, want 0.7", cfg.GetSimilarityThreshold())
	case cfg.GetWebHTTPHost() != "127.0.0.1:8888":
		t.Errorf("GetWebHTTPHost() = %q, want 127.0.0.1:8888", cfg.GetWebHTTPHost())
	case cfg.GetQdrantHost() != "":
		t.Errorf("GetQdrantHost() = %q, want empty", cfg.GetQdrantHost())
	case cfg.GetQdrantAPIKey() != "":
		t.Errorf("GetQdrantAPIKey() = %q, want empty", cfg.GetQdrantAPIKey())
	case cfg.GetUsersJSON() != "":
		t.Errorf("GetUsersJSON() = %q, want empty", cfg.GetUsersJSON())
	case cfg.RequireAuth():
		t.Error("RequireAuth() = true without a Qdrant host; that would pick Qdrant on a default install")
	}
}

// A mandatory setting that is absent or empty must stop startup, and the error
// must name the variable: an operator reads it and sets exactly that one.
//
// `notEmpty` covers both cases, while `required` covers only the first — it is
// satisfied by presence, so `AIMEM_DIMENSIONS=` used to pass parsing and fail at
// the first request with a zero width. Startup is the only place where the
// operator can still act on the message.
func TestGetConfigRejectsMissingOrEmptyMandatorySetting(t *testing.T) {
	t.Parallel()

	mandatory := []string{
		"AIMEM_LLM_HOST",
		"AIMEM_SUMMARIZATION_MODEL",
		"AIMEM_DIMENSIONS",
	}

	cases := map[string]struct {
		value string
		keep  bool
	}{
		"unset": {},
		"empty": {value: "", keep: true},
	}

	for _, name := range mandatory {
		for label, tt := range cases {
			t.Run(name+"/"+label, func(t *testing.T) {
				t.Parallel()

				environ := minimalEnvironment()
				if tt.keep {
					environ[name] = tt.value
				} else {
					delete(environ, name)
				}

				_, err := getConfig(environ)
				if err == nil {
					t.Fatalf("%s %s was accepted; startup would succeed with a setting it cannot use", name, label)
				}

				if !strings.Contains(err.Error(), name) {
					t.Errorf("error does not name the variable: %v", err)
				}
			})
		}
	}
}

// Optional settings are unaffected: an empty value there still falls back to
// its default, so clearing a line in a deployment file is safe. Only the
// mandatory three refuse an empty value.
func TestGetConfigEmptyValueFallsBackToTheDefault(t *testing.T) {
	t.Parallel()

	environ := minimalEnvironment()
	environ["AIMEM_TOP_K"] = ""
	environ["AIMEM_SIMILARITY_THRESHOLD"] = ""
	environ["AIMEM_WEB_HTTP_HOST"] = ""

	cfg, err := getConfig(environ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	switch {
	case cfg.GetTopK() != 5:
		t.Errorf("GetTopK() = %d, want the default 5", cfg.GetTopK())
	case cfg.GetSimilarityThreshold() != 0.7:
		t.Errorf("GetSimilarityThreshold() = %v, want the default 0.7", cfg.GetSimilarityThreshold())
	case cfg.GetWebHTTPHost() != "127.0.0.1:8888":
		t.Errorf("GetWebHTTPHost() = %q, want the default 127.0.0.1:8888", cfg.GetWebHTTPHost())
	}
}

// The threshold is a similarity, so a value outside 0..1 is a mistake in the
// deployment file; it is clamped rather than rejected, because refusing to start
// over a threshold is a worse outcome than a search that returns everything.
func TestGetConfigClampsSimilarityThreshold(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value string
		want  float32
	}{
		"above one":  {"2", 1},
		"below zero": {"-1", 0},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			environ := minimalEnvironment()
			environ["AIMEM_SIMILARITY_THRESHOLD"] = tt.value

			cfg, err := getConfig(environ)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got := cfg.GetSimilarityThreshold(); got != tt.want {
				t.Errorf("GetSimilarityThreshold() = %v, want %v", got, tt.want)
			}
		})
	}
}
