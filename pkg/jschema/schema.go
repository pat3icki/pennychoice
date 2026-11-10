package jschema

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/google/jsonschema-go/jsonschema"
	// "github.com/google/jsonschema-go"
)

// mu protects package-level access to v
var mu sync.RWMutex

// v holds the registry of loaded schemas by a unique name.
var v = make(map[string]*jsonschema.Schema)

var ErrSchemaNotFound = errors.New("schema not found in registry")

// LoadSchema parses a JSON schema (from bytes, string, or direct struct) and adds it to the registry.
func LoadSchema[T *jsonschema.Schema | []byte | string](name string, schemaInput T) error {
	mu.Lock()
	defer mu.Unlock()

	var js *jsonschema.Schema

	// Use a type switch to handle the union type correctly
	switch val := any(schemaInput).(type) {
	case *jsonschema.Schema:
		js = val
	case []byte:
		js = &jsonschema.Schema{}
		if err := json.Unmarshal(val, js); err != nil {
			return fmt.Errorf("failed to unmarshal schema bytes: %w", err)
		}
	case string:
		js = &jsonschema.Schema{}
		if err := json.Unmarshal([]byte(val), js); err != nil {
			return fmt.Errorf("failed to unmarshal schema string: %w", err)
		}
	}

	v[name] = js
	return nil
}

// Validate resolves the stored schema and validates the given payload against it.
func Validate(name string, data []byte, opts *jsonschema.ResolveOptions) error {
	mu.RLock()
	schema, exists := v[name]
	mu.RUnlock()

	if !exists {
		return ErrSchemaNotFound
	}

	// In the google/jsonschema-go package, schemas must be resolved before validating.
	resolved, err := schema.Resolve(opts)
	if err != nil {
		return fmt.Errorf("failed to resolve schema: %w", err)
	}

	// Unmarshal the byte payload into a map/slice/primitive before validating
	var payload any
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal validation payload: %w", err)
	}

	// Validate checks the unmarshaled payload structure against the resolved schema.
	if err := resolved.Validate(payload); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// Generate infers a JSON Schema structure from a generic Go type using reflection.
// This is exposed directly as a package-level function.
func Generate[T any]() (*jsonschema.Schema, error) {
	schema, err := jsonschema.For[T](nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate schema: %w", err)
	}
	return schema, nil
}
