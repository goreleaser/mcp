package main

import (
	"io"
	"reflect"
	"strings"

	"github.com/goreleaser/goreleaser-pro/v2/pkg/config"
	"go.yaml.in/yaml/v4"
)

// parse reads a goreleaser configuration from an io.Reader and returns a config.Project.
func parse(r io.Reader) (*config.Project, error) {
	var proj config.Project
	if err := yaml.NewDecoder(r).Decode(&proj); err != nil {
		return nil, err
	}
	return &proj, nil
}

// findDeprecated returns a map of deprecated fields that have non-zero values.
// The keys are the composed field names (e.g., 'archives.builds', 'brews').
func findDeprecated(cfg *config.Project) map[string]struct{} {
	deprecated := make(map[string]struct{})
	checkDeprecatedFields(reflect.ValueOf(cfg).Elem(), "", deprecated)
	return deprecated
}

func checkDeprecatedFields(v reflect.Value, prefix string, deprecated map[string]struct{}) {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Get the yaml tag name
		yamlTag := fieldType.Tag.Get("yaml")
		if yamlTag == "" || yamlTag == "-" {
			continue
		}
		// Remove omitempty and other options
		yamlName := strings.Split(yamlTag, ",")[0]

		// Build the composed name
		var composedName string
		if prefix == "" {
			composedName = yamlName
		} else {
			composedName = prefix + "." + yamlName
		}

		// Check if field has "deprecated=true" in jsonschema tag
		isDeprecated := strings.Contains(fieldType.Tag.Get("jsonschema"), "deprecated")

		// Check if the field is non-zero
		if isDeprecated && !isZero(field) {
			deprecated[composedName] = struct{}{}
		}

		// Recursively check nested structs
		if field.Kind() == reflect.Struct {
			checkDeprecatedFields(field, composedName, deprecated)
		} else if field.Kind() == reflect.Slice {
			for j := 0; j < field.Len(); j++ {
				elem := field.Index(j)
				if elem.Kind() == reflect.Struct {
					checkDeprecatedFields(elem, composedName, deprecated)
				}
			}
		} else if field.Kind() == reflect.Pointer && !field.IsNil() {
			if field.Elem().Kind() == reflect.Struct {
				checkDeprecatedFields(field.Elem(), composedName, deprecated)
			}
		}
	}
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	case reflect.Struct:
		return v.IsZero()
	}
	return false
}
