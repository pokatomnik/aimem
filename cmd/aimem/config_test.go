package main

import (
	"reflect"
	"testing"
)

func TestRequiredConfigFieldsUseSupportedEnvTags(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(Config{})
	for _, want := range []struct {
		field string
		tag   string
	}{
		{field: "EmbeddingModel", tag: "AIMEM_EMBEDDING_MODEL,required"},
		{field: "SummarizationModel", tag: "AIMEM_SUMMARIZATION_MODEL,required"},
		{field: "Dimensions", tag: "AIMEM_DIMENSIONS,required"},
	} {
		field, ok := typ.FieldByName(want.field)
		if !ok {
			t.Fatalf("Config has no field %q", want.field)
		}
		if got := field.Tag.Get("env"); got != want.tag {
			t.Errorf("Config.%s env tag = %q, want %q", want.field, got, want.tag)
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
		t.Fatalf("Config.Dimensions kind = %s, want uint supported by env v3.5.0", got)
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

func TestIsQdrantConfigured(t *testing.T) {
	t.Parallel()

	emptyConfig := Config{}
	if emptyConfig.IsQdrantConfigured() {
		t.Fatal("IsQdrantConfigured() = true for an empty host")
	}

	configuredConfig := Config{QdrantHost: "http://qdrant.example:6333"}
	if !configuredConfig.IsQdrantConfigured() {
		t.Fatal("IsQdrantConfigured() = false for a configured host")
	}
}
